//revive:disable:package-comments
package apps

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v84/github"

	"github.com/katastroma/phortizo/internal/jwt"
	"github.com/katastroma/phortizo/internal/key"
)

const (
	// JWTClockDrift is the amount of time to backdate the issued-at claim
	// to account for clock drift between the server and GitHub.
	JWTClockDrift = 60 * time.Second

	// JWTMaxLifetime is the maximum lifetime GitHub allows for App JWTs.
	JWTMaxLifetime = 10 * time.Minute
)

// App holds GitHub App credentials.
type App struct {
	clientID   string
	privateKey *rsa.PrivateKey
}

// FromAppParameters creates a GitHub App from a client ID and PEM-encoded
// private key. Returns an error if the key cannot be parsed.
func FromAppParameters(clientID string, privateKeyPEM []byte) (*App, error) {
	pk, err := key.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &App{clientID: clientID, privateKey: pk}, nil
}

// MarshalSecret returns the App's credentials as Secret data entries.
func (a *App) MarshalSecret() map[string][]byte {
	return map[string][]byte{
		"client-id":   []byte(a.clientID),
		"private-key": key.MarshalPrivateKey(a.privateKey),
	}
}

// ExchangeInstallationToken signs a JWT and exchanges it for an installation
// token via the GitHub API.
func (a *App) ExchangeInstallationToken(
	ctx context.Context,
	httpClient *http.Client,
	installationID int64,
) (string, error) {
	if a.privateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}

	now := time.Now()
	signed, err := jwt.Sign(a.clientID, now.Add(-JWTClockDrift), now.Add(JWTMaxLifetime), a.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}

	ghClient := github.NewClient(httpClient).WithAuthToken(signed)
	ghToken, _, err := ghClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", fmt.Errorf("creating installation token: %w", err)
	}

	return ghToken.GetToken(), nil
}
