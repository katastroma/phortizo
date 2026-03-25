//revive:disable:package-comments
package webhook

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/verify"
)

var tracer = otel.Tracer("webhook")

// Webhook receives GitHub push webhooks, verifies their signature, matches
// watch targets, and dispatches matched targets for processing.
type Webhook struct {
	log       *slog.Logger
	registrar registration.Registrar
	onMatch   func(ctx context.Context, m match.Result)
}

// New creates a webhook handler.
func New(
	log *slog.Logger,
	registrar registration.Registrar,
	onMatch func(ctx context.Context, m match.Result),
) *Webhook {
	return &Webhook{
		log:       log,
		registrar: registrar,
		onMatch:   onMatch,
	}
}

// ServeHTTP handles POST /webhook/{id}.
func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	registrationID := r.PathValue("registration_id")
	if registrationID == "" {
		http.Error(w, "missing registration id", http.StatusBadRequest)
		return
	}

	reg, err := h.registrar.GetByID(r.Context(), registrationID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "registration lookup failed", "id", registrationID, "error", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to read request body", "error", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if err = verify.Signature(body, reg.Secret, sig); err != nil {
		h.log.WarnContext(r.Context(), "signature verification failed", "id", registrationID, "error", err)
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	ev, err := event.ParsePush(body)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to parse push event", "error", err)
		http.Error(w, "failed to parse event", http.StatusBadRequest)
		return
	}

	matched := match.Targets(reg.WatchTargets, ev)
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
		h.onMatch(ctx, match.Result{
			Registration: reg,
			Target:       target,
			Event:        ev,
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
