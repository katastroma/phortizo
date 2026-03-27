//revive:disable:package-comments
package auth

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Authenticator resolves authentication for git operations.
type Authenticator interface {
	Authenticate(ctx context.Context, httpClient *http.Client) (transport.AuthMethod, error)
}
