//revive:disable:package-comments
package health

import (
	"context"
)

// Service defines something that can have its health checked
type Service interface {
	// Health returns the health of the service
	Health(context.Context) error
}
