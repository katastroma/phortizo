//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"k8s.io/client-go/kubernetes"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
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
	acquireLease     func(ctx context.Context, namespace, name, leaseID string, replayCount int) error
	resolveAuth      func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error)
	clone            func(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error)
	lookupRenderer   func(fs billy.Filesystem, path string) (string, error)
	verifyLease      func(ctx context.Context, namespace, name, leaseID string) (bool, error)
	stream           func(ctx context.Context, fs billy.Filesystem, path, addr string) error
}

// New creates a RetrieverService handler
func New(
	log *slog.Logger,
	tracer tracing.Tracer,
	k8sClient kubernetes.Interface,
	leaseStaleAfter time.Duration,
	maxReplayAttempts int,
	acquireLease func(ctx context.Context, namespace, name, leaseID string, replayCount int) error,
	resolveAuth func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error),
	clone func(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error),
	lookupRenderer func(fs billy.Filesystem, path string) (string, error),
	verifyLease func(ctx context.Context, namespace, name, leaseID string) (bool, error),
	stream func(ctx context.Context, fs billy.Filesystem, path, addr string) error,
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
		lookupRenderer:   lookupRenderer,
		verifyLease:      verifyLease,
		stream:           stream,
	}
}

// Retrieve handles manual source retrieval dispatch
func (r *Retriever) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	namespace := req.GetNamespace()
	watchTargetID := req.GetWatchTargetId()

	r.log.InfoContext(ctx, "retrieve requested", "namespace", namespace, "watch_target", watchTargetID)

	store := configmap.NewStore(r.k8sClient, namespace)
	target, err := source.Get(ctx, store, watchTargetID)
	if err != nil {
		return nil, fmt.Errorf("reading watch target %s/%s: %w", namespace, watchTargetID, err)
	}

	ctx, tracer := tracing.StartEvent(ctx, tracing.EventTypeManual, namespace)
	target.Process(
		ctx, r.log, tracer, namespace, 0,
		r.acquireLease, r.resolveAuth, r.clone, r.lookupRenderer,
		r.verifyLease, r.stream,
	)

	return &pb.RetrieveResponse{}, nil
}
