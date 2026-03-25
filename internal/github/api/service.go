//revive:disable:package-comments
package api

import (
	"context"
)

// Service allows interaction with the GitHub API
type Service interface {
	CreateInstallationToken(
		ctx context.Context,
		installationID int64,
		// TODO Repos?
		// TODO Permissions?
	) (*string, error)
}
