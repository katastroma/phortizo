//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"

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

	t, err := r.tracer.GetTrace(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying trace %s: %w", eventID, err)
	}

	eventSpans := t[tracing.EventSpanName]
	if len(eventSpans) == 0 {
		return nil, fmt.Errorf("no event span found in trace %s", eventID)
	}

	namespace, ok := eventSpans[0][tracing.TenantAttribute]
	if !ok {
		return nil, fmt.Errorf("missing tenant attribute on event span in trace %s", eventID)
	}

	ctx, otelTracer := tracing.StartEvent(ctx, tracing.EventTypeReplay, namespace)

	for _, attrs := range t[tracing.WatchTargetSpanName] {
		target, err := watchTargetFromAttributes(attrs, eventID)
		if err != nil {
			return nil, err
		}

		if err := r.replayTarget(ctx, otelTracer, eventID, namespace, target); err != nil {
			return nil, err
		}
	}

	return &pb.ReplayResponse{}, nil
}

func (r *Retriever) replayTarget(
	ctx context.Context,
	otelTracer trace.Tracer,
	eventID, namespace string,
	target *source.Target,
) error {
	store := configmap.NewStore(r.k8sClient, namespace)
	leaseState, err := lease.Read(ctx, store, target.Name)
	if err != nil {
		return fmt.Errorf("reading lease for %s/%s: %w", namespace, target.Name, err)
	}

	if leaseState.IsLeaseActive(r.leaseStaleAfter) {
		r.log.InfoContext(ctx, "active lease, skipping replay",
			"event_id", eventID,
			"watch_target", target.Name,
			"active_watch_target_lease_id", leaseState.ID,
		)
		return nil
	}

	replayCount := leaseState.ReplayCount() + 1
	if replayCount > r.maxReplayAttemps {
		r.log.ErrorContext(ctx, "max replay attempts exceeded",
			"event_id", eventID,
			"watch_target", target.Name,
			"replay_count", replayCount,
			"max", r.maxReplayAttemps,
		)
		return status.Errorf(codes.FailedPrecondition,
			"max replay attempts (%d) exceeded for watch target %s in event %s",
			r.maxReplayAttemps, target.Name, eventID)
	}

	r.runner.HandleMatch(ctx, otelTracer, namespace, target, replayCount)
	return nil
}

func watchTargetFromAttributes(attrs tracing.Attributes, eventID string) (*source.Target, error) {
	name, ok := attrs[tracing.WatchTargetNameAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.WatchTargetNameAttribute, eventID)
	}

	repoURL, ok := attrs[tracing.WatchTargetRepoURLAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.WatchTargetRepoURLAttribute, eventID)
	}

	ref, ok := attrs[tracing.WatchTargetRefAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.WatchTargetRefAttribute, eventID)
	}

	path, ok := attrs[tracing.WatchTargetPathAttribute]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in event %s", tracing.WatchTargetPathAttribute, eventID)
	}

	return &source.Target{Name: name, RepoURL: repoURL, Ref: ref, Path: path}, nil
}
