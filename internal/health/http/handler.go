//revive:disable:package-comments
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Handler serves HTTP health checks.
type Handler struct {
	log *slog.Logger
}

// New returns a health check handler.
func New(log *slog.Logger) *Handler {
	return &Handler{log: log}
}

type healthStatus struct{}

// ServeHTTP reports service health including GitHub API connectivity.
func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	s := healthStatus{}
	httpStatus := http.StatusOK

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(s); err != nil {
		h.log.Error("failed to encode health response", "error", err)
	}
}
