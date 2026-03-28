//revive:disable:package-comments
package webhook

import (
	"context"
	"net/http"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
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

	webhookSecret, err := secret.Read(ctx, h.k8sClient, namespace, SecretName, SecretKey)
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

	store := configmap.NewStore(h.k8sClient, namespace)
	watchTargets, err := source.List(ctx, store)
	if err != nil {
		h.fail(ctx, w, "failed to retrieve watch targets", http.StatusNotFound, err, "tenant", namespace)
		return
	}

	matched := match(watchTargets, pushEvent)
	if len(matched) == 0 {
		h.log.InfoContext(ctx, "no watch targets found for event")
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, tracer := tracing.StartEvent(ctx, tracing.EventTypeWebhook, namespace)

	span := trace.SpanFromContext(ctx)
	deliveryID := r.Header.Get("X-GitHub-Delivery")
	span.SetAttributes(
		attribute.String("github.delivery_id", deliveryID),
		attribute.String("github.head_commit", pushEvent.GetAfter()),
	)
	defer span.End()

	for _, target := range matched {
		h.matcher.HandleMatch(ctx, tracer, namespace, target, 0)
	}

	w.WriteHeader(http.StatusAccepted)
}
