//revive:disable:package-comments
package ghinstallationtoken

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/github"
	gh_api "github.com/katastroma/phortizo/internal/github/api"
	gh_apps "github.com/katastroma/phortizo/internal/github/apps"
)

// Credential represents a GitHub App Installation Credential. It
// auto-refreshes the installation token when expired.
type Credential struct {
	ghService gh_api.Service

	app            *gh_apps.App
	installationID int64

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// New returns a Credential from a GitHub App Installation.
func New(app *gh_apps.App, installationID int64, service gh_api.Service) *Credential {
	return &Credential{ghService: service, app: app, installationID: installationID}
}

// Token returns the installation token, refreshing if expired.
func (c *Credential) Token(ctx context.Context) any {
	token, _ := c.ensureToken(ctx)
	return token
}

// Authenticate returns a git transport AuthMethod, refreshing the token if
// expired.
func (c *Credential) Authenticate(ctx context.Context) transport.AuthMethod {
	token, _ := c.ensureToken(ctx)
	return &githttp.BasicAuth{Username: github.TokenUserName, Password: token}
}

// SetAuth applies the installation token to an HTTP request, refreshing if
// expired.
func (c *Credential) SetAuth(req *http.Request) {
	token, _ := c.ensureToken(req.Context())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Credential) ensureToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	token, err := c.ghService.CreateInstallationToken(ctx, c.installationID)
	if err != nil {
		return "", fmt.Errorf("creating installation token: %w", err)
	}

	c.token = *token
	// GitHub installation tokens expire after 1 hour.
	// Refresh 5 minutes early to avoid edge-case expiry during use.
	c.expiresAt = time.Now().Add(55 * time.Minute)

	return c.token, nil
}
