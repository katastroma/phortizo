//revive:disable:package-comments
package auth

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Credential is something that can authenticate into an AuthTransport
type Credential interface {
	// Authenticate for git operations
	Authenticate(context.Context) transport.AuthMethod

	// SetAuth applies authorization to an HTTP request
	SetAuth(req *http.Request)

	// Token returns the token of the credential
	Token(context.Context) any
}
