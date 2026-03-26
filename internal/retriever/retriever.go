//revive:disable:package-comments
package retriever

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/tracequery"
)

// Retriever implements the naukleros RetrieverService gRPC interface
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log              *slog.Logger
	traces           tracequery.Querier
	runner           match.Handler
	k8sClient        kubernetes.Interface
	leaseStaleAfter  time.Duration
	maxReplayAttemps int
}

// New creates a RetrieverService handler
func New(
	log *slog.Logger,
	traces tracequery.Querier,
	runner match.Handler,
	k8sClient kubernetes.Interface,
	leaseStaleAfter time.Duration,
	maxReplayAttempts int,
) *Retriever {
	return &Retriever{
		log:              log,
		traces:           traces,
		runner:           runner,
		k8sClient:        k8sClient,
		leaseStaleAfter:  leaseStaleAfter,
		maxReplayAttemps: maxReplayAttempts,
	}
}

// Retrieve handles manual source retrieval dispatch
func (r *Retriever) Retrieve(_ context.Context, _ *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Retrieve is not yet implemented")
}
