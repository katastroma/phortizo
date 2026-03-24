//revive:disable:package-comments
package memory

import (
	"sync"

	"github.com/katastroma/phortizo/internal/registration"
)

var _ registration.Registrar = (*Registrar)(nil)

// Registrar is an in-memory implementation of registration.Registrar.
type Registrar struct {
	mu            sync.RWMutex
	registrations map[string]*registration.Record
}

// New returns a new in-memory registration store.
func New() *Registrar {
	return &Registrar{registrations: make(map[string]*registration.Record)}
}
