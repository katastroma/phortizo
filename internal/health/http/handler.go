//revive:disable:package-comments
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/katastroma/phortizo/internal/health"
)

// Handler serves HTTP health checks.
type Handler struct {
	log      *slog.Logger
	services []health.Service
}

// New returns a health check handler.
func New(log *slog.Logger, services ...health.Service) *Handler {
	return &Handler{log: log, services: services}
}

type serviceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ServeHTTP reports service health by checking all registered services.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	statuses := make([]serviceStatus, 0, len(h.services))
	httpStatus := http.StatusOK

	for _, svc := range h.services {
		status := "ok"
		if err := svc.Health(ctx); err != nil {
			h.log.ErrorContext(ctx, "health check failed", "service", svc.Name(), "error", err)
			status = "error"
			httpStatus = http.StatusServiceUnavailable
		}

		statuses = append(statuses, serviceStatus{Name: svc.Name(), Status: status})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		h.log.Error("failed to encode health response", "error", err)
	}
}
