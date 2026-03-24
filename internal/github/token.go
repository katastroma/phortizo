//revive:disable:package-comments
package github

import (
	"context"
	"fmt"
	"time"

	gh "github.com/google/go-github/v84/github"
)

// ExchangeInstallationToken exchanges a signed JWT and installation ID
// for an installation access token. Returns the token string and expiry.
func ExchangeInstallationToken(
	ctx context.Context,
	jwtToken string,
	installationID int64,
) (string, time.Time, error) {
	appClient := gh.NewClient(nil).WithAuthToken(jwtToken)

	token, _, err := appClient.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("creating installation token: %w", err)
	}

	return token.GetToken(), token.GetExpiresAt().Time, nil
}
