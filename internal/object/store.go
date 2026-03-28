//revive:disable:package-comments
package object

import "context"

// Store reads, creates, and updates objects.
type Store interface {
	Get(ctx context.Context, name string) (Resource, error)
	Put(ctx context.Context, obj Resource) error
	Update(ctx context.Context, obj Annotatable) error
}
