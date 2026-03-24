//revive:disable:package-comments
package memory

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/registration"
)

// GetByID returns the registration with the given ID.
func (s *Registrar) GetByID(_ context.Context, id string) (*registration.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.registrations[id]
	if !ok {
		return nil, fmt.Errorf("registration %q: not found", id)
	}

	return r, nil
}
