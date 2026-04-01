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
	r.log.InfoContext(ctx, "replay requested", "event_id", eventID)

	r.log.DebugContext(ctx, "querying trace", "event_id", eventID)
	t, err := r.tracer.GetTrace(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying trace %s: %w", eventID, err)
	}
	r.log.DebugContext(ctx, "trace retrieved", "event_id", eventID)

	eventSpans := t[tracing.EventSpanName]
	if len(eventSpans) == 0 {
		return nil, fmt.Errorf("no event span found in trace %s", eventID)
	}

	namespace, ok := eventSpans[0][tracing.TenantAttribute]
	if !ok {
		return nil, fmt.Errorf("missing tenant attribute on event span in trace %s", eventID)
	}

	log := r.log.With("tenant", namespace)
	ctx, otelTracer := tracing.StartEvent(ctx, tracing.EventTypeReplay, namespace)

	for _, attrs := range t[tracing.SourceTargetSpanName] {
		log.DebugContext(ctx, "reconstructing source target from trace")
		target, err := sourceTargetFromAttributes(attrs, eventID)
		if err != nil {
			return nil, err
		}
		log.DebugContext(ctx, "source target reconstructed", "source_target", target.Name)

		if err := r.replayTarget(ctx, log, otelTracer, eventID, namespace, target); err != nil {
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
	log.DebugContext(ctx, "reading lease state", "source_target", target.Name)
	store := configmap.NewStore(r.k8sClient, namespace)
	leaseState, err := lease.Read(ctx, store, target.Name)
	if err != nil {
		return fmt.Errorf("reading lease for %s/%s: %w", namespace, target.Name, err)
	}
	log.DebugContext(ctx, "lease state read", "source_target", target.Name)

	if leaseState.IsLeaseActive(r.leaseStaleAfter) {
		log.InfoContext(ctx, "active lease, skipping replay",
			"event_id", eventID,
			"source_target", target.Name,
			"active_source_target_lease_id", leaseState.ID,
		)
		return nil
	}

	replayCount := leaseState.ReplayCount() + 1
	if replayCount > r.maxReplayAttemps {
		log.ErrorContext(ctx, "max replay attempts exceeded",
			"event_id", eventID,
			"source_target", target.Name,
			"replay_count", replayCount,
			"max", r.maxReplayAttemps,
		)
		return status.Errorf(codes.FailedPrecondition,
			"max replay attempts (%d) exceeded for source target %s in event %s",
			r.maxReplayAttemps, target.Name, eventID)
	}

	log.InfoContext(ctx, "replaying source target", "replay_count", replayCount)
	err = target.Process(
		ctx, log, otelTracer, namespace, replayCount,
		r.acquireLease, r.resolveAuth, r.clone,
		r.verifyLease, r.stream,
	)
	if err != nil {
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
