//revive:disable:package-comments
package match

import (
	"context"

	"github.com/katastroma/phortizo/internal/onboarding"
)

// Handler processes matched watch targets through the pipeline
type Handler interface {
	HandleMatch(
		ctx context.Context,
		namespace string,
		m onboarding.WatchTarget,
		replayCount int,
	)
}
