//revive:disable:package-comments
package auth

import githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

// TokenUserName is the username for GitHub token-based basic auth.
const TokenUserName = "x-access-token"

// FromBasicAuth returns an AuthMethod for git operations using basic auth.
func FromBasicAuth(username, password string) *githttp.BasicAuth {
	return &githttp.BasicAuth{
		Username: username,
		Password: password,
	}
}
