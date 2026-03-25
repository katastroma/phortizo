//revive:disable:package-comments
package client

import (
	"context"

	"github.com/google/go-github/v84/github"
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

// CreateInstallationToken creates an installation token
// for the GitHub App using an installation token
func (s *Service) CreateInstallationToken(
	ctx context.Context,
	installationID int64,
) (*string, error) {
	ghToken, _, err := s.gh.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, err
	}

	token := ghToken.Token
	return token, nil
}
