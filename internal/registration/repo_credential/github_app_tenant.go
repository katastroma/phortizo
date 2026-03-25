//revive:disable:package-comments
package credential

import "github.com/katastroma/phortizo/internal/github/apps"

// GitHubAppTenant uses GitHub App credentials to issue tokens
type GitHubAppTenant struct {
	apps.App
}
