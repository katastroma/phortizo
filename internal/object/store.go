//revive:disable:package-comments
package object

import "context"

// Store reads, lists, creates, and updates objects.
type Store[T string | []byte] interface {
	New(name string) Resource[T]
	Get(ctx context.Context, name string) (Resource[T], error)
	List(ctx context.Context, labels map[string]string) ([]Resource[T], error)
	Put(ctx context.Context, obj Resource[T]) error
	Update(ctx context.Context, obj Annotatable) error
}
