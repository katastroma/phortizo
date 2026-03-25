//revive:disable:package-comments
package testing

import "github.com/katastroma/phortizo/internal/registration"

// TestRegistration returns a registration record for use in tests.
func TestRegistration() *registration.Record {
	return &registration.Record{
		ID:            "reg-1",
		TenantID:      "acme",
		Secret:        []byte("test-secret"),
		CredentialRef: "cred-1",
		WatchTargets: []registration.WatchTarget{
			{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		},
	}
}
