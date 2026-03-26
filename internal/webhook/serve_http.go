//revive:disable:package-comments
package webhook

import (
	"context"
	"net/http"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/match"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) fail(
	ctx context.Context,
	w http.ResponseWriter,
	msg string,
	code int,
	err error,
	attrs ...any,
) {
	args := []any{"error", err}
	args = append(args, attrs...)
	h.log.ErrorContext(ctx, msg, args...)
	http.Error(w, msg, code)
}

// ServeHTTP handles POST /webhook/{namespace}.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespace := r.PathValue("namespace")
	if namespace == "" {
		http.Error(w, "missing namespace", http.StatusBadRequest)
		return
	}

	watchTargets, err := configmap.ListWatchTargets(ctx, h.k8sClient, namespace)
	if err != nil {
		h.fail(ctx, w, "failed to retrieve watch targets", http.StatusNotFound, err, "tenant", namespace)
		return
	}

	webhookSecret, err := secret.ReadWebhookSecret(ctx, h.k8sClient, namespace, secret.WebhookSecretName)
	if err != nil {
		h.fail(ctx, w, "failed to retrieve secret", http.StatusNotFound, err, "tenant", namespace)
		return
	}

	payload, err := github.ValidatePayload(r, webhookSecret)
	if err != nil {
		h.fail(ctx, w, "payload validation failed", http.StatusUnauthorized, err, "tenant", namespace)
		return
	}

	parsed, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		h.fail(ctx, w, "failed to parse webhook", http.StatusBadRequest, err, "tenant", namespace)
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
		h.runner.HandleMatch(ctx, namespace, target, 0)
	}

	w.WriteHeader(http.StatusAccepted)
}
