//revive:disable:package-comments
package credential

import "github.com/katastroma/phortizo/internal/github/apps"

// GitHubAppPlatform uses the Platform GitHub App to issue tokens
type GitHubAppPlatform struct {
	platformApp *apps.App
}
