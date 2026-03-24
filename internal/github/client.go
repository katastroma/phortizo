//revive:disable:package-comments
package github

import (
	"context"
	"log/slog"
	"sync"
	"time"

	gh "github.com/google/go-github/v84/github"
)

const tokenRefreshBuffer = 5 * time.Minute

// Client is a GitHub API client authenticated via an App installation.
// The installation token is cached and refreshed automatically before expiry.
type Client struct {
	log            *slog.Logger
	app            *App
	installationID int64

	mu        sync.RWMutex
	client    *gh.Client
	expiresAt time.Time
}

// NewClientFromAppInstallation creates a Client for an installation of this App.
// The caller only needs to provide an installation ID — the App's own
// credentials handle JWT signing and token exchange.
func NewClientFromAppInstallation(log *slog.Logger, app *App, installationID int64) *Client {
	return &Client{
		log:            log,
		app:            app,
		installationID: installationID,
	}
}

// Get returns a gh.Client with a valid installation token.
// Refreshes the token if it is expired or about to expire.
func (c *Client) Get(ctx context.Context) (*gh.Client, error) {
	c.mu.RLock()
	if c.client != nil && time.Now().Before(c.expiresAt.Add(-tokenRefreshBuffer)) {
		client := c.client
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil && time.Now().Before(c.expiresAt.Add(-tokenRefreshBuffer)) {
		return c.client, nil
	}

	if err := c.refresh(ctx); err != nil {
		return nil, err
	}

	return c.client, nil
}

// Check verifies GitHub API connectivity using this installation's token.
func (c *Client) Check(ctx context.Context) error {
	client, err := c.Get(ctx)
	if err != nil {
		return err
	}

	_, _, err = client.Meta.Get(ctx)
	return err
}

func (c *Client) refresh(ctx context.Context) error {
	jwtToken, err := c.app.JWT()
	if err != nil {
		return err
	}

	token, expiresAt, err := ExchangeInstallationToken(ctx, jwtToken, c.installationID)
	if err != nil {
		return err
	}

	c.client = gh.NewClient(nil).WithAuthToken(token)
	c.expiresAt = expiresAt
	c.log.Info("installation token refreshed",
		"installation_id", c.installationID,
		"expires_at", c.expiresAt,
	)

	return nil
}
