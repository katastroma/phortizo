//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"
	"sync"
)

var _ Store = (*MemoryStore)(nil)

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu          sync.RWMutex
	credentials map[string]*Credential
}

// NewMemoryStore returns a new in-memory credential store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{credentials: make(map[string]*Credential)}
}

// Add inserts a credential. It is intended for test setup.
func (s *MemoryStore) Add(ref string, c *Credential) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[ref] = c
}

// Get returns the credential for the given ref.
func (s *MemoryStore) Get(_ context.Context, ref string) (*Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.credentials[ref]
	if !ok {
		return nil, fmt.Errorf("credential %q: not found", ref)
	}

	return c, nil
}
