//revive:disable:package-comments
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/katastroma/phortizo/internal/credential"
	gh "github.com/katastroma/phortizo/internal/github"
)

// TokenExchanger exchanges a GitHub App installation ID for an access token.
type TokenExchanger interface {
	InstallationToken(ctx context.Context, installationID int64) (string, time.Time, error)
}

// Resolve converts a credential into a git transport AuthMethod.
func Resolve(ctx context.Context, cred *credential.Credential, platformApp TokenExchanger) (transport.AuthMethod, error) {
	switch cred.Type {
	case credential.AppInstallation:
		token, _, err := platformApp.InstallationToken(ctx, cred.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("exchanging app installation token: %w", err)
		}
		return FromBasicAuth(TokenUserName, token), nil

	case credential.TenantApp:
		app := gh.NewApp(cred.ClientID, cred.PrivateKeyPEM)
		token, _, err := app.InstallationToken(ctx, cred.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("exchanging tenant app token: %w", err)
		}
		return FromBasicAuth(TokenUserName, token), nil

	case credential.Token:
		return FromBasicAuth(TokenUserName, cred.Token), nil

	case credential.BasicAuth:
		return FromBasicAuth(cred.Username, cred.Password), nil

	case credential.SSH:
		return FromSSHKey(cred.SSHKeyPEM)

	default:
		return nil, fmt.Errorf("unsupported credential type: %s", cred.Type)
	}
}
