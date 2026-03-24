//revive:disable:package-comments
package github

import (
	"context"
	"crypto/rsa"
	"sync"
	"time"

	"github.com/katastroma/phortizo/internal/jwt"
	"github.com/katastroma/phortizo/internal/key"
)

// App holds the platform's GitHub App credentials. It caches the parsed
// private key and provides methods to create clients and auth methods
// for any installation of this App.
type App struct {
	id            string
	privateKeyPEM []byte

	mu  sync.Mutex
	key *rsa.PrivateKey
}

// NewApp creates a platform GitHub App.
func NewApp(clientID string, privateKeyPEM []byte) *App {
	return &App{
		id:            clientID,
		privateKeyPEM: privateKeyPEM,
	}
}

// JWT returns a signed JWT for this App.
func (a *App) JWT() (string, error) {
	if err := a.ensureKey(); err != nil {
		return "", err
	}

	now := time.Now()
	issuedAt := now.Add(-60 * time.Second) // GitHub recommends the past issuance to avoid clock skew
	expiresAt := now.Add(10 * time.Minute)
	return jwt.Sign(a.id, issuedAt, expiresAt, a.key)
}

// InstallationToken exchanges an installation ID for an access token using the
// App's credentials.
func (a *App) InstallationToken(
	ctx context.Context,
	installationID int64,
) (string, time.Time, error) {
	jwtToken, err := a.JWT()
	if err != nil {
		return "", time.Time{}, err
	}

	return ExchangeInstallationToken(ctx, jwtToken, installationID)
}

func (a *App) ensureKey() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.key != nil {
		return nil
	}

	key, err := key.ParsePrivateKey(a.privateKeyPEM)
	if err != nil {
		return err
	}

	a.key = key
	return nil
}
