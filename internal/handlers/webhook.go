//revive:disable:package-comments
package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/verify"
)

// Webhook receives GitHub push webhooks, verifies their signature, matches
// watch targets, and dispatches matched targets for processing.
type Webhook struct {
	log     *slog.Logger
	store   registration.Store
	onMatch func(m match.Result)
}

// NewWebhook creates a webhook handler.
func NewWebhook(log *slog.Logger, store registration.Store, onMatch func(m match.Result)) *Webhook {
	return &Webhook{
		log:     log,
		store:   store,
		onMatch: onMatch,
	}
}

// ServeHTTP handles POST /webhook/{id}.
func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing registration id", http.StatusBadRequest)
		return
	}

	reg, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.log.ErrorContext(r.Context(), "registration lookup failed", "id", id, "error", err)
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
		h.log.WarnContext(r.Context(), "signature verification failed", "id", id, "error", err)
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

	for _, target := range matched {
		h.onMatch(match.Result{
			Registration: reg,
			Target:       target,
			Event:        ev,
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
