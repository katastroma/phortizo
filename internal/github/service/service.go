//revive:disable:package-comments
package client

import (
	"context"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/github/apps"
)

// Service allows interacting with the GitHub API
type Service struct {
	gh *github.Client
}

// New creates a new Service
func New(gh *github.Client) *Service {
	return &Service{gh}
}

// Health checks the health of the GitHub Service
func (s *Service) Health(ctx context.Context) error {
	_, _, err := s.gh.Meta.Get(ctx)
	return err
}

// func (s *Service) EnsureToken(ctx context.Context) (string, error) {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	if c.token != "" && time.Now().Before(c.expiresAt) {
// 		return c.token, nil
// 	}

// 	// TODO Use app to Sign JWT (from key)
// 	// TODO Use JWT token to create authenticated github client

// 	token, err := s.createInstallationToken(ctx, app)
// 	if err != nil {
// 		return "", fmt.Errorf("creating installation token: %w", err)
// 	}

// 	c.token = *token
// 	// GitHub installation tokens expire after 1 hour.
// 	// Refresh 5 minutes early to avoid edge-case expiry during use.
// 	c.expiresAt = time.Now().Add(55 * time.Minute)

// 	return c.token, nil
// }

// CreateInstallationToken creates an installation token
// for the GitHub App using an installation token
func (s *Service) createInstallationToken(
	ctx context.Context,
	_ *apps.App,
	installationID int64,
) (*string, error) {
	// TODO Use app to Sign JWT (from key)
	// TODO Use JWT token to create authenticated github client

	ghToken, _, err := s.gh.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, err
	}

	token := ghToken.Token
	return token, nil
}
