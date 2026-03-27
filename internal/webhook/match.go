//revive:disable:package-comments
package webhook

import (
	"slices"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/source"
)

// match watch targets with the given push event.
//
// A target matches when its repo URL and ref equal the event's, and at least one
// changed path has the target's path as a prefix.
func match(targets []*source.Target, ev *github.PushEvent) []*source.Target {
	url := ev.GetRepo().GetCloneURL()
	ref := ev.GetRef()
	changed := changedPaths(ev)

	var matched []*source.Target
	for _, t := range targets {
		if matches(t, url, ref, changed) {
			matched = append(matched, t)
		}
	}
	return matched
}

func hasPrefix(path string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, path)
	}
}

func matches(wt *source.Target, url, ref string, paths []string) bool {
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
