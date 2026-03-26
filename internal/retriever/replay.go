//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"

	pb "github.com/katastroma/naukleros"
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

	namespace, ok := attrs["tenant"]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in trace %s", "tenant", runID)
	}

	repoURL, ok := attrs["watch_target.repo_url"]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in trace %s", "watch_target.repo_url", runID)
	}

	ref, ok := attrs["watch_target.ref"]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in trace %s", "watch_target.ref", runID)
	}

	path, ok := attrs["watch_target.path"]
	if !ok {
		return nil, fmt.Errorf("missing attribute %q in trace %s", "watch_target.path", runID)
	}

	target := registration.WatchTarget{RepoURL: repoURL, Ref: ref, Path: path}

	// TODO Replay prevention: query OTel collector for in-flight runs matching
	// this watch target. Skip if a newer run supersedes. Track replay count
	// via span attributes — if over MAX_REPLAY_ATTEMPTS (env var, needs
	// documenting in README), report permanent failure.

	r.runner.HandleMatch(ctx, namespace, target)

	return &pb.ReplayResponse{}, nil
}
