//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	credential1 "github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/github/apps"
	credential "github.com/katastroma/phortizo/internal/onboarding/repo_credential"
)

// Deserializer converts raw Secret data into a credential.
type Deserializer func(data map[string][]byte) (credential1.Authenticator, error)

// Reader reads credential Secrets from tenant namespaces.
type Reader struct {
	client        kubernetes.Interface
	deserializers map[string]Deserializer
}

// NewReader returns a Reader.
func NewReader(client kubernetes.Interface, platformApp *apps.App) *Reader {
	return &Reader{
		client: client,
		deserializers: map[string]Deserializer{
			credential.TypeGitHubToken:     credential.GitHubTokenFromSecret,
			credential.TypeBasicAuth:       credential.BasicAuthFromSecret,
			credential.TypeSSHKey:          credential.SSHKeyFromSecret,
			credential.TypeGitHubAppTenant: credential.GitHubAppTenantFromSecret,
			credential.TypeGitHubAppPlatform: func(data map[string][]byte) (credential1.Authenticator, error) {
				return credential.GitHubAppPlatformFromSecret(data, platformApp)
			},
		},
	}
}

// Get reads a Secret by name from the given namespace and returns the
// deserialized credential.
func (r *Reader) Get(ctx context.Context, namespace, name string) (credential1.Authenticator, error) {
	s, err := r.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading secret %s/%s: %w", namespace, name, err)
	}

	rawType, ok := s.Data["type"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "type")
	}

	fn, ok := r.deserializers[string(rawType)]
	if !ok {
		return nil, fmt.Errorf("unrecognized credential type %q", string(rawType))
	}

	return fn(s.Data)
}
