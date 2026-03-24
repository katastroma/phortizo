//revive:disable:package-comments
package event

import (
	"encoding/json"
	"fmt"
)

// Push is the relevant data extracted from a GitHub push webhook payload.
type Push struct {
	RepoURL      string
	Ref          string
	ChangedPaths []string
}

// ParsePush extracts a Push from a raw GitHub push webhook payload.
func ParsePush(body []byte) (Push, error) {
	var payload githubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return Push{}, fmt.Errorf("unmarshaling push payload: %w", err)
	}

	seen := make(map[string]struct{})
	for _, c := range payload.Commits {
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
		RepoURL:      payload.Repository.CloneURL,
		Ref:          payload.Ref,
		ChangedPaths: paths,
	}, nil
}

// githubPushPayload is the subset of the GitHub push event payload used for
// parsing.
type githubPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	Commits []struct {
		Added    []string `json:"added"`
		Removed  []string `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
}
