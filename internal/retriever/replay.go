//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
)

// Replay a source retrieval from a previous event
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	eventID := req.GetEventId()
	log := r.log.With("event_id", eventID)
	log.InfoContext(ctx, "replay requested")

	log.DebugContext(ctx, "querying trace")
	t, err := r.tracer.GetTrace(ctx, eventID)
	if err != nil {
		log.ErrorContext(ctx, "querying trace failed", "error", err)
		return nil, fmt.Errorf("querying trace %s: %w", eventID, err)
	}
	log.DebugContext(ctx, "trace retrieved")

	log.DebugContext(ctx, "retrieving spans")
	eventSpans := t[tracing.EventSpanName]
	if len(eventSpans) == 0 {
		log.ErrorContext(ctx, "no event span found in trace")
		return nil, fmt.Errorf("no event span found in trace %s", eventID)
	}
	log.DebugContext(ctx, "spans retrieved")

	log.DebugContext(ctx, "retrieving namespace")
	namespace, ok := eventSpans[0][tracing.TenantAttribute]
	if !ok {
		log.ErrorContext(ctx, "missing tenant attribute on event span")
		return nil, fmt.Errorf("missing tenant attribute on event span in trace %s", eventID)
	}
	log = log.With("tenant", namespace)
	log.DebugContext(ctx, "namespace retrieved")

	ctx, otelTracer, eventSpan := tracing.StartEvent(ctx, tracing.EventTypeReplay, namespace)
	defer eventSpan.End()

	for _, attrs := range t[tracing.SourceTargetSpanName] {
		log.DebugContext(ctx, "reconstructing source target from trace")
		target, err := sourceTargetFromAttributes(attrs, eventID)
		if err != nil {
			log.ErrorContext(ctx, "reconstructing source target failed", "error", err)
			return nil, err
		}
		log = log.With(
			"source_target", target.Name,
			"repo", target.RepoURL,
			"ref", target.Ref,
			"path", target.Path,
		)
		log.DebugContext(ctx, "source target reconstructed")

		err = r.replayTarget(ctx, log, otelTracer, eventID, namespace, target)
		if err != nil {
			log.ErrorContext(ctx, "replay target failed", "error", err)
			return nil, err
		}
	}

	return &pb.ReplayResponse{}, nil
}

func (r *Retriever) replayTarget(
	ctx context.Context,
	log *slog.Logger,
	otelTracer trace.Tracer,
	eventID, namespace string,
	target *source.Target,
) error {
	log.DebugContext(ctx, "check lease")
	store := configmap.NewStore(r.k8sClient, namespace)
	leaseState, err := lease.Read(ctx, store, target.Name)
	if err != nil {
		log.ErrorContext(ctx, "reading lease failed", "error", err)
		return fmt.Errorf("reading lease for %s/%s: %w", namespace, target.Name, err)
	}
	log.DebugContext(ctx, "lease checked")

	log.DebugContext(ctx, "reading lease state")
	if leaseState.IsLeaseActive(r.leaseStaleAfter) {
		log.InfoContext(ctx, "active lease, skipping replay",
			"active_source_target_lease_id", leaseState.ID,
		)
		return nil
	}
	log.DebugContext(ctx, "lease state verified")

	replayCount := leaseState.ReplayCount() + 1
	log = log.With("replay_count", replayCount)

	if replayCount > r.maxReplayAttemps {
		log.ErrorContext(ctx, "max replay attempts exceeded",
			"source_target", target.Name,
			"max", r.maxReplayAttemps,
		)
		return status.Errorf(codes.FailedPrecondition,
			"max replay attempts (%d) exceeded for source target %s in event %s",
			r.maxReplayAttemps, target.Name, eventID)
	}

	log.InfoContext(ctx, "replaying source target")
	err = target.Process(
		ctx, log, otelTracer, namespace, replayCount,
		r.acquireLease, r.resolveAuth, r.clone,
		r.verifyLease, r.stream,
	)
	if err != nil {
		log.ErrorContext(ctx, "replaying source target failed", "error", err)
		return fmt.Errorf("replaying source target failed: %w", err)
	}

	return nil
}

func sourceTargetFromAttributes(attrs tracing.Attributes, eventID string) (*source.Target, error) {
	name, ok := attrs[tracing.SourceTargetNameAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.SourceTargetNameAttribute, eventID)
	}

	repoURL, ok := attrs[tracing.SourceTargetRepoURLAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.SourceTargetRepoURLAttribute, eventID)
	}

	ref, ok := attrs[tracing.SourceTargetRefAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.SourceTargetRefAttribute, eventID)
	}

	path, ok := attrs[tracing.SourceTargetPathAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.SourceTargetPathAttribute, eventID)
	}

	return &source.Target{Name: name, RepoURL: repoURL, Ref: ref, Path: path}, nil
}
