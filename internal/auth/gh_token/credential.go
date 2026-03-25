//revive:disable:package-comments
package ghtoken

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/github"
)

// Credential represents a GitHub App Installation Credential
type Credential struct {
	token string
}

// New returns a new Credential with the token set
func New(token string) *Credential {
	return &Credential{token}
}

// Token returns the token of the credential
func (c *Credential) Token(context.Context) any {
	return c.token
}

// Authenticate returns a token ready for git operations
func (c *Credential) Authenticate(context.Context) transport.AuthMethod {
	return &githttp.BasicAuth{Username: github.TokenUserName, Password: c.token}
}

// SetAuth sets up authentication for an http request using a GitHub App Installation Credential
func (c *Credential) SetAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
}
