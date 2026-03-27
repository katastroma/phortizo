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

// TypeGitHubAppTenant is the Secret type for tenant GitHub App credentials.
const TypeGitHubAppTenant = "github-app-tenant"

// GitHubAppTenant authenticates using a tenant-owned GitHub App.
type GitHubAppTenant struct {
	app            *apps.App
	installationID int64
}

// NewGitHubAppTenant returns a GitHubAppTenant credential.
func NewGitHubAppTenant(app *apps.App, installationID int64) *GitHubAppTenant {
	return &GitHubAppTenant{app: app, installationID: installationID}
}

// GitHubAppTenantFromSecret deserializes a GitHubAppTenant from Secret data.
func GitHubAppTenantFromSecret(data map[string][]byte) (auth.Authenticator, error) {
	clientID, ok := data["client-id"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "client-id")
	}

	pk, ok := data["private-key"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "private-key")
	}

	rawID, ok := data["installation-id"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "installation-id")
	}

	installationID, err := strconv.ParseInt(string(rawID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing installation-id: %w", err)
	}

	app, err := apps.FromAppParameters(string(clientID), pk)
	if err != nil {
		return nil, fmt.Errorf("parsing tenant app credentials: %w", err)
	}

	return NewGitHubAppTenant(app, installationID), nil
}

// MarshalSecret serializes the credential to Secret data.
func (c *GitHubAppTenant) MarshalSecret() map[string][]byte {
	appData := c.app.MarshalSecret()
	return map[string][]byte{
		"type":            []byte(TypeGitHubAppTenant),
		"client-id":       appData["client-id"],
		"private-key":     appData["private-key"],
		"installation-id": []byte(strconv.FormatInt(c.installationID, 10)),
	}
}

// Authenticate signs a JWT, exchanges it for an installation token, and
// returns a BasicAuth transport.
func (c *GitHubAppTenant) Authenticate(ctx context.Context, httpClient *http.Client) (transport.AuthMethod, error) {
	token, err := c.app.ExchangeInstallationToken(ctx, httpClient, c.installationID)
	if err != nil {
		return nil, err
	}

	return &githttp.BasicAuth{Username: gh.TokenUserName, Password: token}, nil
}
