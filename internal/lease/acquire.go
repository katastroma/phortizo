//revive:disable:package-comments
package lease

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/katastroma/phortizo/internal/object"
)

// AcquireFunc acquires a lease on a named object in a namespace.
type AcquireFunc func(ctx context.Context, namespace, name, leaseID string, replayCount int) error

// NewAcquireFunc returns an AcquireFunc that uses the given store factory.
func NewAcquireFunc(newStore object.StoreFactory[string]) AcquireFunc {
	return func(ctx context.Context, namespace, name, leaseID string, replayCount int) error {
		return Acquire(ctx, newStore(namespace), name, leaseID, replayCount)
	}
}

// Acquire attempts to take ownership of an Object, using optimistic concurrency via resourceVersion
// to prevent races
func Acquire(
	ctx context.Context,
	s object.Store[string],
	name string,
	leaseID string,
	replayCount int,
) error {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[IDAnnotation] = leaseID
	annotations[StartedAnnotation] = time.Now().UTC().Format(time.RFC3339)
	annotations[ReplayCountAnnotation] = strconv.Itoa(replayCount)
	obj.SetAnnotations(annotations)

	if err = s.Update(ctx, obj); err != nil {
		return fmt.Errorf("acquiring lease on %s: %w", name, err)
	}

	return nil
}
