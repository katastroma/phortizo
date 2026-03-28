//revive:disable:package-comments
package webhook

import (
	"context"
	"log/slog"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"k8s.io/client-go/kubernetes"
)

// Handler receives GitHub push webhooks, verifies their signature, matches
// watch targets, and dispatches matched targets for processing.
type Handler struct {
	log            *slog.Logger
	k8sClient      kubernetes.Interface
	acquireLease   func(ctx context.Context, namespace, name, leaseID string, replayCount int) error
	resolveAuth    func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error)
	clone          func(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error)
	lookupRenderer func(fs billy.Filesystem, path string) (string, error)
	verifyLease    func(ctx context.Context, namespace, name, leaseID string) (bool, error)
	stream         func(ctx context.Context, fs billy.Filesystem, path, addr string) error
}

// New creates a webhook handler.
func New(
	log *slog.Logger,
	k8sClient kubernetes.Interface,
	acquireLease func(ctx context.Context, namespace, name, leaseID string, replayCount int) error,
	resolveAuth func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error),
	clone func(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error),
	lookupRenderer func(fs billy.Filesystem, path string) (string, error),
	verifyLease func(ctx context.Context, namespace, name, leaseID string) (bool, error),
	stream func(ctx context.Context, fs billy.Filesystem, path, addr string) error,
) *Handler {
	return &Handler{
		log:            log,
		k8sClient:      k8sClient,
		acquireLease:   acquireLease,
		resolveAuth:    resolveAuth,
		clone:          clone,
		lookupRenderer: lookupRenderer,
		verifyLease:    verifyLease,
		stream:         stream,
	}
}
