//revive:disable:package-comments
package tests

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// TestCredential is a credential that should be used for testing
type TestCredential struct {
	token string
}

// NewTestCredential creates a new test credential with the token set
func NewTestCredential(token string) *TestCredential {
	return &TestCredential{token}
}

// Token returns the token of the credential
func (c *TestCredential) Token(context.Context) (any, error) {
	return c.token, nil
}

// Authenticate returns auth ready for git operations
func (c *TestCredential) Authenticate(context.Context) (transport.AuthMethod, error) {
	t := &githttp.BasicAuth{Username: "test", Password: c.token}
	return t, nil
}

// SetAuth sets up authentication for an http request using a GitHub App Installation Credential
func (c *TestCredential) SetAuth(req *http.Request) (err error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return
}
