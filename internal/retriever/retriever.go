//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"k8s.io/client-go/kubernetes"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
)

// Retriever implements the naukleros RetrieverService gRPC interface
type Retriever struct {
	pb.UnimplementedRetrieverServiceServer
	log              *slog.Logger
	tracer           tracing.Tracer
	k8sClient        kubernetes.Interface
	leaseStaleAfter  time.Duration
	maxReplayAttemps int
	acquireLease     lease.AcquireFunc
	resolveAuth      credential.ResolveFunc
	clone            git.CloneFunc
	verifyLease      lease.VerifyFunc
	stream           renderer.StreamFunc
}

// New creates a RetrieverService handler
func New(
	log *slog.Logger,
	tracer tracing.Tracer,
	k8sClient kubernetes.Interface,
	leaseStaleAfter time.Duration,
	maxReplayAttempts int,
	acquireLease lease.AcquireFunc,
	resolveAuth credential.ResolveFunc,
	clone git.CloneFunc,
	verifyLease lease.VerifyFunc,
	stream renderer.StreamFunc,
) *Retriever {
	return &Retriever{
		log:              log,
		tracer:           tracer,
		k8sClient:        k8sClient,
		leaseStaleAfter:  leaseStaleAfter,
		maxReplayAttemps: maxReplayAttempts,
		acquireLease:     acquireLease,
		resolveAuth:      resolveAuth,
		clone:            clone,
		verifyLease:      verifyLease,
		stream:           stream,
	}
}

// Retrieve handles manual source retrieval dispatch
func (r *Retriever) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	namespace := req.GetNamespace()
	sourceTargetID := req.GetSourceTargetId()
	log := r.log.With("tenant", namespace)

	log.InfoContext(ctx, "retrieve requested", "source_target", sourceTargetID)

	log.DebugContext(ctx, "getting source target", "source_target", sourceTargetID)
	store := configmap.NewStore(r.k8sClient, namespace)
	target, err := source.Get(ctx, store, sourceTargetID)
	if err != nil {
		return nil, fmt.Errorf("reading source target %s/%s: %w", namespace, sourceTargetID, err)
	}
	log.DebugContext(ctx, "source target retrieved", "source_target", sourceTargetID)

	log.InfoContext(ctx, "processing source target", "source_target", sourceTargetID)
	ctx, tracer := tracing.StartEvent(ctx, tracing.EventTypeManual, namespace)
	err = target.Process(
		ctx, log, tracer, namespace, 0,
		r.acquireLease, r.resolveAuth, r.clone,
		r.verifyLease, r.stream,
	)
	if err != nil {
		return nil, fmt.Errorf("processing source target failed: %w", err)
	}

	return &pb.RetrieveResponse{}, nil
}
