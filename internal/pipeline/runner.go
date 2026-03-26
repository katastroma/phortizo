//revive:disable:package-comments
package pipeline

import (
	"log/slog"

	"go.opentelemetry.io/otel"

	"github.com/katastroma/phortizo/internal/source"
)

var tracer = otel.Tracer("pipeline")

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log       *slog.Logger
	renderers map[source.RendererType]string
}

// New creates a pipeline runner.
func New(log *slog.Logger, renderers map[source.RendererType]string) *Runner {
	return &Runner{log: log, renderers: renderers}
}
