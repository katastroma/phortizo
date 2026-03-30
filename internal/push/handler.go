//revive:disable:package-comments
package push

import (
	"log/slog"

	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/renderer"
)

// Handler receives GitHub push webhooks, verifies their signature, matches
// source targets, and dispatches matched targets for processing.
type Handler struct {
	log          *slog.Logger
	k8sClient    kubernetes.Interface
	acquireLease lease.AcquireFunc
	resolveAuth  credential.ResolveFunc
	clone        git.CloneFunc
	verifyLease  lease.VerifyFunc
	stream       renderer.StreamFunc
}

// New creates a webhook handler.
func New(
	log *slog.Logger,
	k8sClient kubernetes.Interface,
	acquireLease lease.AcquireFunc,
	resolveAuth credential.ResolveFunc,
	clone git.CloneFunc,
	verifyLease lease.VerifyFunc,
	stream renderer.StreamFunc,
) *Handler {
	return &Handler{
		log:          log,
		k8sClient:    k8sClient,
		acquireLease: acquireLease,
		resolveAuth:  resolveAuth,
		clone:        clone,
		verifyLease:  verifyLease,
		stream:       stream,
	}
}
