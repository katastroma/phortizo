//revive:disable:package-comments
package transport

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/katastroma/phortizo/internal/github/apps"
)

const (
	// InstallationTokenLifetime is the duration a GitHub App installation
	// token is valid.
	InstallationTokenLifetime = 60 * time.Minute

	// InstallationTokenRefreshBuffer is the amount of time before expiry
	// to refresh the token, avoiding edge-case expiry during use.
	InstallationTokenRefreshBuffer = 5 * time.Minute
)

// InstallationTokenAuth is an http.RoundTripper that authenticates requests
// using a GitHub App installation token. It caches the token and refreshes
// it automatically before expiry.
type InstallationTokenAuth struct {
	app            *apps.App
	installationID int64
	base           http.RoundTripper

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// NewInstallationTokenAuth returns an InstallationTokenAuth that authenticates
// using the given App and installation ID. The base RoundTripper is used for
// both token exchange and authenticated requests. If base is nil,
// http.DefaultTransport is used.
func NewInstallationTokenAuth(
	app *apps.App, installationID int64, base http.RoundTripper,
) *InstallationTokenAuth {
	if base == nil {
		base = http.DefaultTransport
	}

	return &InstallationTokenAuth{
		app:            app,
		installationID: installationID,
		base:           base,
	}
}

// RoundTrip ensures a valid installation token is set on the request.
func (t *InstallationTokenAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.ensureToken(req)
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return t.base.RoundTrip(req)
}

func (t *InstallationTokenAuth) ensureToken(req *http.Request) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token != "" && time.Now().Before(t.expiresAt) {
		return t.token, nil
	}

	exchangeClient := &http.Client{Transport: t.base}
	token, err := t.app.ExchangeInstallationToken(req.Context(), exchangeClient, t.installationID)
	if err != nil {
		return "", fmt.Errorf("refreshing installation token: %w", err)
	}

	t.token = token
	t.expiresAt = time.Now().Add(InstallationTokenLifetime - InstallationTokenRefreshBuffer)

	return t.token, nil
}
