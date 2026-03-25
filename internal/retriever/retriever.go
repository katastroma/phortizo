//revive:disable:package-comments
package retriever

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/tracequery"
)

// Retriever implements the naukleros RetrieverService gRPC interface
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log    *slog.Logger
	traces tracequery.Querier
	runner match.Handler
}

// New creates a RetrieverService handler
func New(
	log *slog.Logger,
	traces tracequery.Querier,
	runner match.Handler,
) *Retriever {
	return &Retriever{
		log:    log,
		traces: traces,
		runner: runner,
	}
}

// Retrieve handles manual source retrieval dispatch
func (r *Retriever) Retrieve(_ context.Context, _ *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Retrieve is not yet implemented")
}

// Replay a source retrieval from a previous run
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	runID := req.GetRunId()
	r.log.InfoContext(ctx, "replay requested", "run_id", runID)

	// TODO Get tenant ID and watch targets from previous trace run ID
	var tenantID string
	var result match.Result

	// TODO Replay prevention: query OTel collector for in-flight runs matching
	// this watch target. Skip if a newer run supersedes. Track replay count
	// via span attributes — if over MAX_REPLAY_ATTEMPTS (env var, needs
	// documenting in README), report permanent failure.

	r.runner.HandleMatch(ctx, tenantID, result)

	return &pb.ReplayResponse{}, nil
}
