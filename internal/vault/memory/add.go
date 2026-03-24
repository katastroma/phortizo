//revive:disable:package-comments
package memory

import "github.com/katastroma/phortizo/internal/vault"

// Add inserts a credential. It is intended for test setup.
func (s *Vault) Add(ref string, c *vault.Credential) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[ref] = c
}
