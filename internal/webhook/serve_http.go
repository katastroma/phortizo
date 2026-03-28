//revive:disable:package-comments
package webhook

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
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

	secretStore := secret.NewStore(h.k8sClient, namespace)
	resource, err := secretStore.Get(ctx, SecretName)
	if err != nil {
		h.fail(ctx, w, "failed to retrieve secret", http.StatusNotFound, err, "tenant", namespace)
		return
	}

	webhookSecret, found := resource.GetData()[SecretKey]
	if !found {
		msg := "secret key not found"
		err = fmt.Errorf("missing key %q in secret %s/%s", SecretKey, namespace, SecretName)
		h.fail(ctx, w, msg, http.StatusNotFound, err, "tenant", namespace)
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
		target.Process(
			ctx, h.log, tracer, namespace, 0,
			h.acquireLease, h.resolveAuth, h.clone, h.lookupRenderer,
			h.verifyLease, h.stream,
		)
	}

	w.WriteHeader(http.StatusAccepted)
}
