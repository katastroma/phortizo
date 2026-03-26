//revive:disable:package-comments
package webhook

import (
	"log/slog"
	"net/http"

	"github.com/google/go-github/v84/github"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/match"
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

	ctx := r.Context()

	watchTargets, err := configmap.ListWatchTargets(ctx, h.k8sClient, namespace)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to retrieve watch targets", "error", err)
		http.Error(w, "failed to retrieve watch targets", http.StatusNotFound)
		return
	}

	webhookSecret, err := secret.ReadWebhookSecret(ctx, h.k8sClient, namespace, secret.WebhookSecretName)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to retrieve secret", "error", err)
		http.Error(w, "failed to retrieve secret", http.StatusNotFound)
		return
	}

	payload, err := github.ValidatePayload(r, webhookSecret)
	if err != nil {
		h.log.WarnContext(ctx, "payload validation failed", "id", namespace, "error", err)
		http.Error(w, "payload validation failed", http.StatusUnauthorized)
		return
	}

	parsed, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		h.log.ErrorContext(ctx, "failed to parse webhook", "error", err)
		http.Error(w, "failed to parse webhook", http.StatusBadRequest)
		return
	}

	pushEvent, ok := parsed.(*github.PushEvent)
	if !ok {
		h.log.WarnContext(ctx, "ignored non-push event", "type", github.WebHookType(r))
		w.WriteHeader(http.StatusOK)
		return
	}

	matched := match.Find(watchTargets, pushEvent)
	if len(matched) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	deliveryID := r.Header.Get("X-GitHub-Delivery")
	ctx, span := tracer.Start(ctx, "webhook.dispatch", trace.WithAttributes(
		attribute.String("github.delivery_id", deliveryID),
		attribute.String("github.head_commit", pushEvent.GetAfter()),
	))
	defer span.End()

	for _, target := range matched {
		h.runner.HandleMatch(ctx, namespace, match.Result{
			Target: target,
			Event:  pushEvent,
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
