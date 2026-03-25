//revive:disable:package-comments
package github

import "github.com/katastroma/phortizo/internal/auth/resolve"

var _ resolve.ExchangerFactory = (*AppFactory)(nil)

// AppFactory creates TokenExchangers from GitHub App credentials.
type AppFactory struct{}

// Create returns a new App configured with the given credentials.
func (f *AppFactory) Create(clientID string, privateKeyPEM []byte) resolve.TokenExchanger {
	return NewApp(clientID, privateKeyPEM)
}
