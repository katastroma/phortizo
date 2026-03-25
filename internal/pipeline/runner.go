//revive:disable:package-comments
package pipeline

import (
	"log/slog"

	"go.opentelemetry.io/otel"

	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/vault"
)

var tracer = otel.Tracer("pipeline")

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log   *slog.Logger
	vault vault.Withdrawer
	// ghClient  gh_api.Client
	renderers map[source.RendererType]string
}

// New creates a pipeline runner.
func New(
	log *slog.Logger,
	vault vault.Withdrawer,
	// ghClient gh_api.Client,
	renderers map[source.RendererType]string,
) *Runner {
	return &Runner{
		log:   log,
		vault: vault,
		// ghClient:  ghClient,
		renderers: renderers,
	}
}
