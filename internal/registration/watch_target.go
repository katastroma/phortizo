//revive:disable:package-comments
package registration

// WatchTarget represents a tracked location within a repository.
type WatchTarget struct {
	RepoURL string
	Ref     string
	Path    string
}
