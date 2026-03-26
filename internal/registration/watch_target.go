//revive:disable:package-comments
package registration

import "fmt"

// WatchTarget represents a tracked location within a repository.
type WatchTarget struct {
	RepoURL   string
	Ref       string
	Path      string
	Overrides []byte
}

// WatchTargetFromConfigMap deserializes a WatchTarget from ConfigMap data.
func WatchTargetFromConfigMap(data map[string]string) (WatchTarget, error) {
	repoURL, ok := data["repo-url"]
	if !ok {
		return WatchTarget{}, fmt.Errorf("missing key %q", "repo-url")
	}

	ref, ok := data["ref"]
	if !ok {
		return WatchTarget{}, fmt.Errorf("missing key %q", "ref")
	}

	path, ok := data["path"]
	if !ok {
		return WatchTarget{}, fmt.Errorf("missing key %q", "path")
	}

	var overrides []byte
	if raw, ok := data["overrides"]; ok {
		overrides = []byte(raw)
	}

	return WatchTarget{
		RepoURL:   repoURL,
		Ref:       ref,
		Path:      path,
		Overrides: overrides,
	}, nil
}

// MarshalConfigMap serializes the WatchTarget to ConfigMap data.
func (t WatchTarget) MarshalConfigMap() map[string]string {
	data := map[string]string{
		"repo-url": t.RepoURL,
		"ref":      t.Ref,
		"path":     t.Path,
	}

	if len(t.Overrides) > 0 {
		data["overrides"] = string(t.Overrides)
	}

	return data
}
