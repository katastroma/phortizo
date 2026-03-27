//revive:disable:package-comments
package match

import (
	"context"

	"github.com/katastroma/phortizo/internal/source"
	"go.opentelemetry.io/otel/trace"
)

// Handler processes matched watch targets through the pipeline
type Handler interface {
	HandleMatch(
		ctx context.Context,
		tracer trace.Tracer,
		namespace string,
		m *source.Target,
		replayCount int,
	)
}
