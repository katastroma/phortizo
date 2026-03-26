//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/auth"
	"github.com/katastroma/phortizo/internal/github/apps"
	credential "github.com/katastroma/phortizo/internal/onboarding/repo_credential"
)

// Reader reads credential Secrets from tenant namespaces.
type Reader struct {
	client      kubernetes.Interface
	platformApp *apps.App
}

// NewReader returns a Reader.
func NewReader(client kubernetes.Interface, platformApp *apps.App) *Reader {
	return &Reader{client: client, platformApp: platformApp}
}

// Get reads a Secret by name from the given namespace and returns the
// deserialized credential.
func (r *Reader) Get(ctx context.Context, namespace, name string) (auth.Credential, error) {
	s, err := r.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading secret %s/%s: %w", namespace, name, err)
	}

	return Deserialize(s.Data, r.platformApp)
}

// Deserialize converts Secret data into a credential based on the type field.
func Deserialize(data map[string][]byte, platformApp *apps.App) (auth.Credential, error) {
	rawType, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "type")
	}

	switch string(rawType) {
	case credential.TypeGitHubToken:
		return credential.GitHubTokenFromSecret(data)

	case credential.TypeBasicAuth:
		return credential.BasicAuthFromSecret(data)

	case credential.TypeSSHKey:
		return credential.SSHKeyFromSecret(data)

	case credential.TypeGitHubAppTenant:
		return credential.GitHubAppTenantFromSecret(data)

	case credential.TypeGitHubAppPlatform:
		return credential.GitHubAppPlatformFromSecret(data, platformApp)

	default:
		return nil, fmt.Errorf("unrecognized credential type %q", string(rawType))
	}
}
