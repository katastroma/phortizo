//revive:disable:package-comments
package match

import (
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/registration"
)

// Find returns the watch targets that match the given push event. A target
// matches when its repo URL and ref equal the event's, and at least one
// changed path has the target's path as a prefix.
func Find(targets []registration.WatchTarget, ev *github.PushEvent) []registration.WatchTarget {
	changed := changedPaths(ev)

	var matched []registration.WatchTarget
	for _, t := range targets {
		if t.RepoURL != ev.GetRepo().GetCloneURL() || t.Ref != ev.GetRef() {
			continue
		}

		for _, p := range changed {
			if strings.HasPrefix(p, t.Path) {
				matched = append(matched, t)
				break
			}
		}
	}
	return matched
}

// changedPaths collects and deduplicates all file paths affected across every
// commit in the push event.
func changedPaths(ev *github.PushEvent) []string {
	seen := make(map[string]struct{})
	for _, c := range ev.Commits {
		for _, p := range c.Added {
			seen[p] = struct{}{}
		}
		for _, p := range c.Removed {
			seen[p] = struct{}{}
		}
		for _, p := range c.Modified {
			seen[p] = struct{}{}
		}
	}

	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}
	return paths
}
