//revive:disable:package-comments
package credential

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/katastroma/phortizo/internal/object"
)

// Authenticator resolves authentication for git operations.
type Authenticator interface {
	Authenticate(ctx context.Context, httpClient *http.Client) (transport.AuthMethod, error)
}

// Reader reads credentials from a store.
type Reader interface {
	Get(ctx context.Context, store object.Store[[]byte], name string) (Authenticator, error)
}
