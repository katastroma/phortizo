//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"k8s.io/client-go/kubernetes"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/tracing"
)

// Retriever implements the naukleros RetrieverService gRPC interface
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log              *slog.Logger
	tracer           tracing.Tracer
	runner           match.Handler
	k8sClient        kubernetes.Interface
	leaseStaleAfter  time.Duration
	maxReplayAttemps int
}

// New creates a RetrieverService handler
func New(
	log *slog.Logger,
	tracer tracing.Tracer,
	runner match.Handler,
	k8sClient kubernetes.Interface,
	leaseStaleAfter time.Duration,
	maxReplayAttempts int,
) *Retriever {
	return &Retriever{
		log:              log,
		tracer:           tracer,
		runner:           runner,
		k8sClient:        k8sClient,
		leaseStaleAfter:  leaseStaleAfter,
		maxReplayAttemps: maxReplayAttempts,
	}
}

// Retrieve handles manual source retrieval dispatch
func (r *Retriever) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	namespace := req.GetNamespace()
	watchTargetID := req.GetWatchTargetId()

	r.log.InfoContext(ctx, "retrieve requested", "namespace", namespace, "watch_target", watchTargetID)

	target, err := configmap.GetWatchTarget(ctx, r.k8sClient, namespace, watchTargetID)
	if err != nil {
		return nil, fmt.Errorf("reading watch target %s/%s: %w", namespace, watchTargetID, err)
	}

	ctx, tracer := tracing.StartEvent(ctx, tracing.EventTypeManual, namespace)
	r.runner.HandleMatch(ctx, tracer, namespace, target, 0)

	return &pb.RetrieveResponse{}, nil
}
