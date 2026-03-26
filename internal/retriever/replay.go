//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/pipeline"
	"github.com/katastroma/phortizo/internal/registration"
)

// Replay a source retrieval from a previous run
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	runID := req.GetRunId()
	r.log.InfoContext(ctx, "replay requested", "run_id", runID)

	attrs, err := r.traces.SpanAttributes(ctx, runID, pipeline.SpanName)
	if err != nil {
		return nil, fmt.Errorf("querying trace %s: %w", runID, err)
	}

	namespace, target, err := watchTargetFromAttributes(attrs, runID)
	if err != nil {
		return nil, err
	}

	lease, err := configmap.ReadLease(ctx, r.k8sClient, namespace, target.Name)
	if err != nil {
		return nil, fmt.Errorf("reading lease for %s/%s: %w", namespace, target.Name, err)
	}

	if lease.IsLeaseActive(r.leaseStaleAfter) {
		r.log.InfoContext(ctx, "active lease, skipping replay",
			"run_id", runID,
			"active_run_id", lease.RunID,
		)
		return &pb.ReplayResponse{}, nil
	}

	replayCount := lease.ReplayCount + 1
	if replayCount > r.maxReplayAttemps {
		r.log.ErrorContext(ctx, "max replay attempts exceeded",
			"run_id", runID,
			"replay_count", replayCount,
			"max", r.maxReplayAttemps,
		)
		return nil, status.Errorf(codes.FailedPrecondition,
			"max replay attempts (%d) exceeded for run %s", r.maxReplayAttemps, runID)
	}

	r.runner.HandleMatch(ctx, namespace, target, replayCount)

	return &pb.ReplayResponse{}, nil
}

func watchTargetFromAttributes(attrs map[string]string, runID string) (string, registration.WatchTarget, error) {
	namespace, ok := attrs["tenant"]
	if !ok {
		return "", registration.WatchTarget{}, fmt.Errorf("missing attribute %q in trace %s", "tenant", runID)
	}

	name, ok := attrs["watch_target.name"]
	if !ok {
		return "", registration.WatchTarget{}, fmt.Errorf("missing attribute %q in trace %s", "watch_target.name", runID)
	}

	repoURL, ok := attrs["watch_target.repo_url"]
	if !ok {
		return "", registration.WatchTarget{}, fmt.Errorf("missing attribute %q in trace %s", "watch_target.repo_url", runID)
	}

	ref, ok := attrs["watch_target.ref"]
	if !ok {
		return "", registration.WatchTarget{}, fmt.Errorf("missing attribute %q in trace %s", "watch_target.ref", runID)
	}

	path, ok := attrs["watch_target.path"]
	if !ok {
		return "", registration.WatchTarget{}, fmt.Errorf("missing attribute %q in trace %s", "watch_target.path", runID)
	}

	return namespace, registration.WatchTarget{
		Name:    name,
		RepoURL: repoURL,
		Ref:     ref,
		Path:    path,
	}, nil
}
