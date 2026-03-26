//revive:disable:package-comments
package retriever

import (
	"context"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/match"
)

// Replay a source retrieval from a previous run
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	runID := req.GetRunId()
	r.log.InfoContext(ctx, "replay requested", "run_id", runID)

	// TODO Get tenant namespace and watch targets from previous trace run ID
	var namespace string
	var result match.Result

	// TODO Replay prevention: query OTel collector for in-flight runs matching
	// this watch target. Skip if a newer run supersedes. Track replay count
	// via span attributes — if over MAX_REPLAY_ATTEMPTS (env var, needs
	// documenting in README), report permanent failure.

	r.runner.HandleMatch(ctx, namespace, result)

	return &pb.ReplayResponse{}, nil
}
