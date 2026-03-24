//revive:disable:package-comments
package registration

import (
	"context"
	"fmt"
	"sync"
)

var _ Store = (*MemoryStore)(nil)

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu            sync.RWMutex
	registrations map[string]*Registration
}

// NewMemoryStore returns a new in-memory registration store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{registrations: make(map[string]*Registration)}
}

// Add inserts a registration. It is intended for test setup.
func (s *MemoryStore) Add(r *Registration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations[r.ID] = r
}

// GetByID returns the registration with the given ID.
func (s *MemoryStore) GetByID(_ context.Context, id string) (*Registration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.registrations[id]
	if !ok {
		return nil, fmt.Errorf("registration %q: not found", id)
	}

	return r, nil
}
