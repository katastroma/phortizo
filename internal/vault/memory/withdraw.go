//revive:disable:package-comments
package memory

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/auth"
)

// Withdraw returns the credential for the given ref.
func (s *Vault) Withdraw(_ context.Context, ref string) (auth.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.credentials[ref]
	if !ok {
		return nil, fmt.Errorf("credential %q: not found", ref)
	}

	return c, nil
}
