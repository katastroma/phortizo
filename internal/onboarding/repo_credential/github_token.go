//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/github"
)

// TypeGitHubToken is the Secret type for GitHub token credentials.
const TypeGitHubToken = "github-token"

// GitHubToken authenticates with a GitHub personal or fine-grained token.
type GitHubToken struct {
	token string
}

// NewGitHubToken returns a GitHubToken credential.
func NewGitHubToken(token string) *GitHubToken {
	return &GitHubToken{token: token}
}

// GitHubTokenFromSecret deserializes a GitHubToken from Secret data.
func GitHubTokenFromSecret(data map[string][]byte) (credential.Authenticator, error) {
	token, ok := data["token"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "token")
	}

	return NewGitHubToken(string(token)), nil
}

// MarshalSecret serializes the credential to Secret data.
func (c *GitHubToken) MarshalSecret() map[string][]byte {
	return map[string][]byte{
		"type":  []byte(TypeGitHubToken),
		"token": []byte(c.token),
	}
}

// Authenticate returns a BasicAuth transport using the token.
func (c *GitHubToken) Authenticate(context.Context, *http.Client) (transport.AuthMethod, error) {
	return &githttp.BasicAuth{Username: github.TokenUserName, Password: c.token}, nil
}
