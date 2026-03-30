//revive:disable:package-comments
package credential

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/katastroma/phortizo/internal/object"
)

// ResolveFunc resolves git transport authentication for a named credential
// secret. Returns nil auth for empty secret names (public repositories).
type ResolveFunc func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error)

// NewResolveFunc returns a ResolveFunc that uses the given reader and store factory.
func NewResolveFunc(
	reader Reader, httpClient *http.Client, newStore object.StoreFactory[[]byte],
) ResolveFunc {
	return func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error) {
		if credentialSecret == "" {
			return nil, nil
		}

		cred, err := reader.Get(ctx, newStore(namespace), credentialSecret)
		if err != nil {
			return nil, err
		}

		return cred.Authenticate(ctx, httpClient)
	}
}
