//revive:disable:package-comments
package registration

// Record is a webhook endpoint configuration for a tenant. It holds the
// shared secret, watch targets, and a reference to the tenant's credentials.
type Record struct {
	ID            string
	TenantID      string
	Secret        []byte
	WatchTargets  []WatchTarget
	CredentialRef string
}
