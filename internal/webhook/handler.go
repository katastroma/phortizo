//revive:disable:package-comments
package webhook

import (
	"log/slog"

	"k8s.io/client-go/kubernetes"

	match_pkg "github.com/katastroma/phortizo/internal/match"
)

// Handler receives GitHub push webhooks, verifies their signature, matches
// watch targets, and dispatches matched targets for processing.
type Handler struct {
	log       *slog.Logger
	matcher   match_pkg.Handler
	k8sClient kubernetes.Interface
}

// New creates a webhook handler.
func New(log *slog.Logger, runner match_pkg.Handler, k8sClient kubernetes.Interface) *Handler {
	return &Handler{log: log, matcher: runner, k8sClient: k8sClient}
}
