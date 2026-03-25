//revive:disable:package-comments
package match

import (
	"context"
)

// Handler processes matched watch targets through the pipeline
type Handler interface {
	HandleMatch(ctx context.Context, tenantID string, m Result)
}
