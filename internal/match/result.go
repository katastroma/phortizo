//revive:disable:package-comments
package match

import (
	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/registration"
)

// Result carries a matched watch target alongside the push event that
// produced the match.
type Result struct {
	Target registration.WatchTarget
	Event  *github.PushEvent
}
