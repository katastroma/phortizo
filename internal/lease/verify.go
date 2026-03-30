//revive:disable:package-comments
package lease

import (
	"context"

	"github.com/katastroma/phortizo/internal/object"
)

// VerifyFunc checks whether a lease is still held by the given ID.
type VerifyFunc func(ctx context.Context, namespace, name, leaseID string) (bool, error)

// NewVerifyFunc returns a VerifyFunc that uses the given store factory.
func NewVerifyFunc(newStore object.StoreFactory[string]) VerifyFunc {
	return func(ctx context.Context, namespace, name, leaseID string) (bool, error) {
		return HeldBy(ctx, newStore(namespace), name, leaseID)
	}
}

// HeldBy checks if the given lease ID currently holds the lease.
func HeldBy(ctx context.Context, s object.Store[string], name string, leaseID string) (bool, error) {
	l, err := Read(ctx, s, name)
	if err != nil {
		return false, err
	}

	return l.ID == leaseID, nil
}
