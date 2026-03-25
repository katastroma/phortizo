//revive:disable:package-comments
package memory

import (
	"github.com/katastroma/phortizo/internal/auth"
)

// Add inserts a credential. It is intended for test setup.
func (s *Vault) Add(ref string, c auth.Credential) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[ref] = c
}
