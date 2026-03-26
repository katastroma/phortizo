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

// Name returns the service name for health reporting.
func (s *Service) Name() string {
	return "github"
}

// Health checks the health of the GitHub Service
func (s *Service) Health(ctx context.Context) error {
	_, _, err := s.gh.Meta.Get(ctx)
	return err
}
