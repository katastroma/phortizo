//revive:disable:package-comments
package webhook

import (
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/verify"
)

var tracer = otel.Tracer("webhook")

// Webhook receives GitHub push webhooks, verifies their signature, matches
// watch targets, and dispatches matched targets for processing.
type Webhook struct {
	log       *slog.Logger
	runner    match.Handler
	k8sClient kubernetes.Interface
}

// New creates a webhook handler.
func New(log *slog.Logger, runner match.Handler, k8sClient kubernetes.Interface) *Webhook {
	return &Webhook{log: log, runner: runner, k8sClient: k8sClient}
}

// ServeHTTP handles POST /webhook/{namespace}.
func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	if namespace == "" {
		http.Error(w, "missing namespace", http.StatusBadRequest)
		return
	}

	// TODO Get tenant watch config(s) from k8s
	var watchTargets []registration.WatchTarget

	ctx := r.Context()
	webhookSecret, err := secret.ReadWebhookSecret(ctx, h.k8sClient, namespace, secret.WebhookSecretName)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to retrieve secret", "error", err)
		http.Error(w, "failed to retrieve secret", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to read request body", "error", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if err = verify.Signature(body, webhookSecret, sig); err != nil {
		h.log.WarnContext(r.Context(), "signature verification failed", "id", namespace, "error", err)
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	ev, err := event.ParsePush(body)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to parse push event", "error", err)
		http.Error(w, "failed to parse event", http.StatusBadRequest)
		return
	}

	matched := match.Find(watchTargets, ev)
	if len(matched) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	deliveryID := r.Header.Get("X-GitHub-Delivery")
	ctx, span := tracer.Start(r.Context(), "webhook.dispatch", trace.WithAttributes(
		attribute.String("github.delivery_id", deliveryID),
	))
	defer span.End()

	for _, target := range matched {
		h.runner.HandleMatch(ctx, namespace, match.Result{
			Target: target,
			Event:  ev,
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
