//revive:disable:package-comments
package memory

import (
	"sync"

	"github.com/katastroma/phortizo/internal/vault"
)

var _ vault.Withdrawer = (*Vault)(nil)

// Vault is an in-memory implementation of credential.Vault.
type Vault struct {
	mu          sync.RWMutex
	credentials map[string]*vault.Credential
}

// New returns a new in-memory credential store.
func New() *Vault {
	return &Vault{credentials: make(map[string]*vault.Credential)}
}
