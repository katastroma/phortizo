//revive:disable:package-comments
package memory

import "github.com/katastroma/phortizo/internal/registration"

// Add inserts a registration. It is intended for test setup.
func (s *Registrar) Add(r *registration.Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations[r.ID] = r
}
