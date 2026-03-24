//revive:disable:package-comments
package registration

import "context"

// WatchTarget represents a tracked location within a repository.
type WatchTarget struct {
	RepoURL string
	Ref     string
	Path    string
}

// Registration is a webhook endpoint configuration for a tenant. It holds the
// shared secret, watch targets, and a reference to the tenant's credentials.
type Registration struct {
	ID            string
	TenantID      string
	Secret        []byte
	WatchTargets  []WatchTarget
	CredentialRef string
}

// Store retrieves registrations.
type Store interface {
	GetByID(ctx context.Context, id string) (*Registration, error)
}
