//revive:disable:package-comments
package object

import "context"

// Store reads, lists, creates, and updates objects.
type Store interface {
	New(name string) Resource
	Get(ctx context.Context, name string) (Resource, error)
	List(ctx context.Context, labels map[string]string) ([]Resource, error)
	Put(ctx context.Context, obj Resource) error
	Update(ctx context.Context, obj Annotatable) error
}
