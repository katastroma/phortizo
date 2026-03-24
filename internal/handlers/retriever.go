//revive:disable:package-comments
package handler

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/katastroma/naukleros"
)

// Retriever implements the naukleros RetrieverService gRPC interface.
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log *slog.Logger
}

// NewRetriever creates a RetrieverService handler.
func NewRetriever(log *slog.Logger) *Retriever {
	return &Retriever{log: log}
}

// Retrieve handles manual source retrieval dispatch.
func (r *Retriever) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Retrieve is not yet implemented")
}

// Replay replays a source retrieval from a previous run.
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	r.log.InfoContext(ctx, "replay requested", "run_id", req.GetRunId())

	// TODO: query trace backend for the original run's span attributes,
	// reconstruct the watch target, check replay prevention, and re-run
	// the pipeline.

	return nil, status.Error(codes.Unimplemented, "Replay is not yet implemented")
}
