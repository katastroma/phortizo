//revive:disable:package-comments
package health

import "context"

// Checker defines something that can have its health checked
type Checker interface {
	// Name() string

	// Check the health
	Check(context.Context) error
}
