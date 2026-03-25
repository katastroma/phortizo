//revive:disable:package-comments
package handler

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/tracequery"
)

// Retriever implements the naukleros RetrieverService gRPC interface.
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log       *slog.Logger
	traces    tracequery.Querier
	registrar registration.Registrar
	onMatch   func(ctx context.Context, m match.Result)
}

// NewRetriever creates a RetrieverService handler.
func NewRetriever(
	log *slog.Logger,
	traces tracequery.Querier,
	registrar registration.Registrar,
	onMatch func(ctx context.Context, m match.Result),
) *Retriever {
	return &Retriever{
		log:       log,
		traces:    traces,
		registrar: registrar,
		onMatch:   onMatch,
	}
}

// Retrieve handles manual source retrieval dispatch.
func (r *Retriever) Retrieve(_ context.Context, _ *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Retrieve is not yet implemented")
}

// Replay replays a source retrieval from a previous run.
func (r *Retriever) Replay(ctx context.Context, req *pb.ReplayRequest) (*pb.ReplayResponse, error) {
	runID := req.GetRunId()
	r.log.InfoContext(ctx, "replay requested", "run_id", runID)

	attrs, err := r.traces.SpanAttributes(ctx, runID, "pipeline.run")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "querying trace: %v", err)
	}

	result, err := r.resultFromTrace(ctx, attrs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "reconstructing run: %v", err)
	}

	r.onMatch(ctx, result)

	return &pb.ReplayResponse{}, nil
}

func (r *Retriever) resultFromTrace(
	ctx context.Context,
	attrs tracequery.Attributes,
) (match.Result, error) {
	regID := attrs["registration_id"]
	if regID == "" {
		return match.Result{}, fmt.Errorf("missing registration_id attribute")
	}

	repoURL := attrs["watch_target.repo_url"]
	ref := attrs["watch_target.ref"]
	path := attrs["watch_target.path"]
	if repoURL == "" || ref == "" {
		return match.Result{}, fmt.Errorf("missing watch target attributes")
	}

	rec, err := r.registrar.GetByID(ctx, regID)
	if err != nil {
		return match.Result{}, fmt.Errorf("looking up registration %s: %w", regID, err)
	}

	return match.Result{
		Registration: rec,
		Target: registration.WatchTarget{
			RepoURL: repoURL,
			Ref:     ref,
			Path:    path,
		},
	}, nil
}
