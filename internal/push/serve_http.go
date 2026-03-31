//revive:disable:package-comments
package push

import (
	"context"
	"fmt"
	"log/slog"
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
	log *slog.Logger,
	w http.ResponseWriter,
	msg string,
	code int,
	err error,
	attrs ...any,
) {
	args := []any{"error", err}
	args = append(args, attrs...)
	log.ErrorContext(ctx, msg, args...)
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

	log := h.log.With("tenant", namespace)
	log.InfoContext(ctx, "push received", "delivery_id", r.Header.Get("X-GitHub-Delivery"))

	secretStore := secret.NewStore(h.k8sClient, namespace)

	log.DebugContext(ctx, "getting secret", "secret", SecretName)
	resource, err := secretStore.Get(ctx, SecretName)
	if err != nil {
		h.fail(ctx, log, w, "failed to retrieve secret", http.StatusNotFound, err)
		return
	}
	log.DebugContext(ctx, "secret retrieved")

	log.DebugContext(ctx, "getting secret data", "secret-key", SecretKey)
	webhookSecret, found := resource.GetData()[SecretKey]
	if !found {
		msg := "secret key not found"
		err = fmt.Errorf("missing key %q in secret %s/%s", SecretKey, namespace, SecretName)
		h.fail(ctx, log, w, msg, http.StatusNotFound, err)
		return
	}
	log.DebugContext(ctx, "secret data retrieved")

	log.DebugContext(ctx, "validating payload")
	payload, err := github.ValidatePayload(r, webhookSecret)
	if err != nil {
		h.fail(ctx, log, w, "payload validation failed", http.StatusUnauthorized, err)
		return
	}
	log.DebugContext(ctx, "payload validated")

	log.DebugContext(ctx, "parsing event")
	parsed, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		h.fail(ctx, log, w, "failed to parse webhook", http.StatusBadRequest, err)
		return
	}
	log.DebugContext(ctx, "event parsed")

	pushEvent, ok := parsed.(*github.PushEvent)
	if !ok {
		log.WarnContext(ctx, "ignored non-push event", "type", github.WebHookType(r))
		w.WriteHeader(http.StatusOK)
		return
	}

	store := configmap.NewStore(h.k8sClient, namespace)

	log.DebugContext(ctx, "getting source targets")
	sourceTargets, err := source.List(ctx, store)
	if err != nil {
		h.fail(ctx, log, w, "failed to retrieve source targets", http.StatusNotFound, err)
		return
	}
	log.DebugContext(ctx, "source targets retrieved")

	log.DebugContext(ctx, "resolving source target matches")
	matched := match(sourceTargets, pushEvent)
	if len(matched) == 0 {
		log.WarnContext(ctx, "no source targets matched")
		w.WriteHeader(http.StatusOK)
		return
	}
	log.DebugContext(ctx, "source targets matched", "count", len(matched))

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
			ctx, log, tracer, namespace, 0,
			h.acquireLease, h.resolveAuth, h.clone,
			h.verifyLease, h.stream,
		)
	}

	w.WriteHeader(http.StatusAccepted)
}
