//revive:disable:package-comments
package lease

import (
	"context"

	"github.com/katastroma/phortizo/internal/object"
)

// VerifyFunc returns a closure that checks whether a lease is still held by the given ID
func VerifyFunc(
	newStore func(string) object.Store[string],
) func(ctx context.Context, namespace, name, leaseID string) (bool, error) {
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
