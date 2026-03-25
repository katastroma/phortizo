//revive:disable:package-comments
package apps

import (
	"crypto/rsa"
	"sync"

	"github.com/katastroma/phortizo/internal/key"
)

// App holds GitHub App credentials. It caches the parsed
// private key and provides methods to create clients using auth methods
// for any installation of this App.
type App struct {
	mu         sync.Mutex
	clientID   string
	privateKey *rsa.PrivateKey
}

// FromAppParameters creates a platform GitHub App using app parameters
func FromAppParameters(clientID string, privateKeyPEM []byte) (*App, error) {
	privateKey, err := key.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &App{clientID: clientID, privateKey: privateKey}, nil
}
