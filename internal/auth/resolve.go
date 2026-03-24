//revive:disable:package-comments
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"

	gh "github.com/katastroma/phortizo/internal/github"
	"github.com/katastroma/phortizo/internal/vault"
)

// TokenExchanger exchanges a GitHub App installation ID for an access token.
type TokenExchanger interface {
	InstallationToken(ctx context.Context, installationID int64) (string, time.Time, error)
}

// Resolve converts a credential into a git transport AuthMethod.
func Resolve(ctx context.Context, cred *vault.Credential, platformApp TokenExchanger) (transport.AuthMethod, error) {
	switch cred.Type {
	case vault.AppInstallation:
		token, _, err := platformApp.InstallationToken(ctx, cred.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("exchanging app installation token: %w", err)
		}
		return FromBasicAuth(TokenUserName, token), nil

	case vault.TenantApp:
		app := gh.NewApp(cred.ClientID, cred.PrivateKeyPEM)
		token, _, err := app.InstallationToken(ctx, cred.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("exchanging tenant app token: %w", err)
		}
		return FromBasicAuth(TokenUserName, token), nil

	case vault.Token:
		return FromBasicAuth(TokenUserName, cred.Token), nil

	case vault.BasicAuth:
		return FromBasicAuth(cred.Username, cred.Password), nil

	case vault.SSH:
		return FromSSHKey(cred.SSHKeyPEM)

	default:
		return nil, fmt.Errorf("unsupported credential type: %s", cred.Type)
	}
}
