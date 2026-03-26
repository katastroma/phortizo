//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/auth"
	gh "github.com/katastroma/phortizo/internal/github"
	"github.com/katastroma/phortizo/internal/github/apps"
)

// TypeGitHubAppPlatform is the Secret type for platform GitHub App credentials.
const TypeGitHubAppPlatform = "github-app-platform"

// GitHubAppPlatform authenticates using the platform GitHub App with a
// tenant-specific installation ID.
type GitHubAppPlatform struct {
	platformApp    *apps.App
	installationID int64
}

// NewGitHubAppPlatform returns a GitHubAppPlatform credential.
func NewGitHubAppPlatform(platformApp *apps.App, installationID int64) *GitHubAppPlatform {
	return &GitHubAppPlatform{platformApp: platformApp, installationID: installationID}
}

// GitHubAppPlatformFromSecret deserializes a GitHubAppPlatform from Secret
// data. The platform App is provided externally since the Secret only stores
// the installation ID.
func GitHubAppPlatformFromSecret(data map[string][]byte, platformApp *apps.App) (auth.Credential, error) {
	rawID, ok := data["installation-id"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "installation-id")
	}

	installationID, err := strconv.ParseInt(string(rawID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing installation-id: %w", err)
	}

	return NewGitHubAppPlatform(platformApp, installationID), nil
}

// MarshalSecret serializes the credential to Secret data.
func (c *GitHubAppPlatform) MarshalSecret() map[string][]byte {
	return map[string][]byte{
		"type":            []byte(TypeGitHubAppPlatform),
		"installation-id": []byte(strconv.FormatInt(c.installationID, 10)),
	}
}

// Authenticate signs a JWT using the platform App, exchanges it for an
// installation token, and returns a BasicAuth transport.
func (c *GitHubAppPlatform) Authenticate(ctx context.Context, httpClient *http.Client) (transport.AuthMethod, error) {
	token, err := c.platformApp.ExchangeInstallationToken(ctx, httpClient, c.installationID)
	if err != nil {
		return nil, err
	}

	return &githttp.BasicAuth{Username: gh.TokenUserName, Password: token}, nil
}
