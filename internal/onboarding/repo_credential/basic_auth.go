//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/auth"
)

// TypeBasicAuth is the Secret type for basic auth credentials.
const TypeBasicAuth = "basic-auth"

// BasicAuth authenticates with a username and password.
type BasicAuth struct {
	username string
	password string
}

// NewBasicAuth returns a BasicAuth credential.
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{username: username, password: password}
}

// BasicAuthFromSecret deserializes a BasicAuth from Secret data.
func BasicAuthFromSecret(data map[string][]byte) (auth.Authenticator, error) {
	username, ok := data["username"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "username")
	}

	password, ok := data["password"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "password")
	}

	return NewBasicAuth(string(username), string(password)), nil
}

// MarshalSecret serializes the credential to Secret data.
func (c *BasicAuth) MarshalSecret() map[string][]byte {
	return map[string][]byte{
		"type":     []byte(TypeBasicAuth),
		"username": []byte(c.username),
		"password": []byte(c.password),
	}
}

// Authenticate returns a BasicAuth transport.
func (c *BasicAuth) Authenticate(context.Context, *http.Client) (transport.AuthMethod, error) {
	return &githttp.BasicAuth{Username: c.username, Password: c.password}, nil
}
