//revive:disable:package-comments
package match

import (
	"strings"

	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/registration"
)

// Find returns the watch targets that match the given push event. A target
// matches when its repo URL and ref equal the event's, and at least one
// changed path has the target's path as a prefix.
func Find(targets []registration.WatchTarget, ev event.Push) []registration.WatchTarget {
	var matched []registration.WatchTarget
	for _, t := range targets {
		if t.RepoURL != ev.RepoURL || t.Ref != ev.Ref {
			continue
		}

		for _, changed := range ev.ChangedPaths {
			if strings.HasPrefix(changed, t.Path) {
				matched = append(matched, t)
				break
			}
		}
	}
	return matched
}
