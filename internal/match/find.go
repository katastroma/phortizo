//revive:disable:package-comments
package match

import (
	"slices"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/registration"
)

func hasPrefix(path string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, path)
	}
}

func matches(wt registration.WatchTarget, url, ref string, paths []string) bool {
	if wt.RepoURL != url || wt.Ref != ref {
		return false
	}

	return slices.ContainsFunc(paths, hasPrefix(wt.Path))
}

// changedPaths collects and deduplicates all file paths affected across every
// commit in the push event.
func changedPaths(ev *github.PushEvent) []string {
	seen := make(map[string]struct{})

	for _, c := range ev.Commits {
		all := slices.Concat(c.Added, c.Modified, c.Removed)
		for _, p := range all {
			seen[p] = struct{}{}
		}
	}

	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}

	return paths
}

// Find returns the watch targets that match the given push event. A target
// matches when its repo URL and ref equal the event's, and at least one
// changed path has the target's path as a prefix.
func Find(targets []registration.WatchTarget, ev *github.PushEvent) []registration.WatchTarget {
	url := ev.GetRepo().GetCloneURL()
	ref := ev.GetRef()
	changed := changedPaths(ev)

	var matched []registration.WatchTarget
	for _, t := range targets {
		if matches(t, url, ref, changed) {
			matched = append(matched, t)
		}
	}
	return matched
}
