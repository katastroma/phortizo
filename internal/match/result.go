//revive:disable:package-comments
package match

import (
	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/registration"
)

// Result carries a matched watch target alongside the registration and event
// that produced the match.
type Result struct {
	Registration *registration.Record
	Target       registration.WatchTarget
	Event        event.Push
}
