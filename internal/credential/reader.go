//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/object"
)

// Deserializer converts raw Secret data into a credential.
type Deserializer func(data map[string][]byte) (Authenticator, error)

type reader struct {
	deserializers map[string]Deserializer
}

// NewReader returns a Reader that deserializes credentials from a store.
func NewReader(platformApp *apps.App) Reader {
	return &reader{
		deserializers: map[string]Deserializer{
			TypeGitHubToken:     GitHubTokenFromSecret,
			TypeBasicAuth:       BasicAuthFromSecret,
			TypeSSHKey:          SSHKeyFromSecret,
			TypeGitHubAppTenant: GitHubAppTenantFromSecret,
			TypeGitHubAppPlatform: func(data map[string][]byte) (Authenticator, error) {
				return GitHubAppPlatformFromSecret(data, platformApp)
			},
		},
	}
}

// Get reads a Secret by name from the store and returns the deserialized
// credential.
func (r *reader) Get(ctx context.Context, store object.Store[[]byte], name string) (Authenticator, error) {
	resource, err := store.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	data := resource.GetData()

	rawType, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "type")
	}

	fn, ok := r.deserializers[string(rawType)]
	if !ok {
		return nil, fmt.Errorf("unrecognized credential type %q", string(rawType))
	}

	return fn(data)
}
