//revive:disable:package-comments
package event

import (
	"github.com/google/go-github/v84/github"
)

// Push is the relevant data extracted from a GitHub push webhook payload.
type Push struct {
	RepoURL      string
	Ref          string
	CommitSHA    string
	ChangedPaths []string
}

// FromPushEvent extracts a Push from a GitHub SDK PushEvent, deduplicating
// changed paths across all commits.
func FromPushEvent(ev *github.PushEvent) Push {
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

	return Push{
		RepoURL:      ev.GetRepo().GetCloneURL(),
		Ref:          ev.GetRef(),
		CommitSHA:    ev.GetAfter(),
		ChangedPaths: paths,
	}
}
