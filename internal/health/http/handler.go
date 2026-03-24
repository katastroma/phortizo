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
	log     *slog.Logger
	checker health.Checker
}

// New returns a health check handler.
func New(log *slog.Logger, checker health.Checker) *Handler {
	return &Handler{log: log, checker: checker}
}

type healthStatus struct {
	GitHub string `json:"github"`
}

// ServeHTTP reports service health including GitHub API connectivity.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := healthStatus{GitHub: "ok"}
	httpStatus := http.StatusOK

	ctx := r.Context()
	if err := h.checker.Check(ctx); err != nil {
		h.log.Error("failed to check upstream health", "error", err)
		s.GitHub = err.Error()
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(s); err != nil {
		h.log.Error("failed to encode health response", "error", err)
	}
}
