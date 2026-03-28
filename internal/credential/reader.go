//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/k8s/secret"
)

// Deserializer converts raw Secret data into a credential.
type Deserializer func(data map[string][]byte) (Authenticator, error)

type reader struct {
	client        kubernetes.Interface
	deserializers map[string]Deserializer
}

// NewReader returns a Reader that deserializes credentials from K8s Secrets.
func NewReader(client kubernetes.Interface, platformApp *apps.App) Reader {
	return &reader{
		client: client,
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

// Get reads a Secret by name from the given namespace and returns the
// deserialized credential.
func (r *reader) Get(ctx context.Context, namespace, name string) (Authenticator, error) {
	data, err := secret.Read(ctx, r.client, namespace, name)
	if err != nil {
		return nil, err
	}

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
