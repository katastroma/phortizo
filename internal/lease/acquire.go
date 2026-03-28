//revive:disable:package-comments
package lease

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/katastroma/phortizo/internal/object"
)

// AcquireFunc returns a closure that acquires a lease on a named object in the given namespace
func AcquireFunc(
	newStore func(string) object.Store[string],
) func(ctx context.Context, namespace, name, leaseID string, replayCount int) error {
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
