//revive:disable:package-comments
package pipeline

import (
	"context"

	"github.com/katastroma/phortizo/internal/match"
)

// Handler processes matched watch targets through the pipeline.
type Handler interface {
	HandleMatch(ctx context.Context, m match.Result)
}
