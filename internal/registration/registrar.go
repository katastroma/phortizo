//revive:disable:package-comments
package registration

import "context"

// Registrar retrieves registrations.
type Registrar interface {
	GetByID(ctx context.Context, id string) (*Record, error)
}
