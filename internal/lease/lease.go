// Package lease controls operations for processing resource ownership.
//
// Kubernetes provides a first-class Lease resource (coordination.k8s.io/v1)
// for leader election and distributed locking. We use annotations instead
// because our objects have external lifecycles. Their lifecycles are managed by
// tenant onboarding, not the processor. Storing lease state as annotations
// avoids creating and garbage-collecting a separate Lease resource per watch
// target per processing instance. The object's resourceVersion provides
// optimistic concurrency for free.
package lease

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Annotatable is something with annotations.
type Annotatable interface {
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

// ObjectStore reads and updates annotatable objects.
type ObjectStore interface {
	Get(ctx context.Context, name string) (Annotatable, error)
	Update(ctx context.Context, obj Annotatable) error
}

const (
	// IDAnnotation identifies which processing instance currently owns the object
	IDAnnotation = "katastroma.org/lease-id"

	// StartedAnnotation is the timestamp when the lease was acquired
	StartedAnnotation = "katastroma.org/lease-started"

	// ReplayCountAnnotation tracks how many times this event has been replayed for this object
	ReplayCountAnnotation = "katastroma.org/lease-replay-count"
)

// State represents the current lease state on an object.
type State struct {
	ID          string
	Started     time.Time
	replayCount int
}

// ReplayCount returns the replay count. Returns 0 on nil receiver.
func (s *State) ReplayCount() int {
	if s == nil {
		return 0
	}

	return s.replayCount
}

// Acquire attempts to take ownership of an Object, using optimistic concurrency
// via resourceVersion to prevent races.
func Acquire(
	ctx context.Context,
	store ObjectStore,
	name string,
	leaseID string,
	replayCount int,
) error {
	obj, err := store.Get(ctx, name)
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

	if err = store.Update(ctx, obj); err != nil {
		return fmt.Errorf("acquiring lease on %s: %w", name, err)
	}

	return nil
}

// Read reads the current lease state from an object.
func Read(ctx context.Context, store ObjectStore, name string) (*State, error) {
	obj, err := store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}

	id := obj.GetAnnotations()[IDAnnotation]
	if id == "" {
		return nil, nil
	}

	started, err := time.Parse(time.RFC3339, obj.GetAnnotations()[StartedAnnotation])
	if err != nil {
		return nil, err
	}

	replayCount, _ := strconv.Atoi(obj.GetAnnotations()[ReplayCountAnnotation])

	return &State{ID: id, Started: started, replayCount: replayCount}, nil
}

// Release clears the lease annotations on an object.
func Release(ctx context.Context, store ObjectStore, name string) error {
	obj, err := store.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}

	annotations := obj.GetAnnotations()
	delete(annotations, IDAnnotation)
	delete(annotations, StartedAnnotation)
	delete(annotations, ReplayCountAnnotation)
	obj.SetAnnotations(annotations)

	if err = store.Update(ctx, obj); err != nil {
		return fmt.Errorf("releasing lease on %s: %w", name, err)
	}

	return nil
}

// HeldBy checks if the given lease ID currently holds the lease.
func HeldBy(ctx context.Context, store ObjectStore, name string, leaseID string) (bool, error) {
	l, err := Read(ctx, store, name)
	if err != nil {
		return false, err
	}

	return l.ID == leaseID, nil
}

// IsLeaseActive checks if a lease is held and not stale.
func (s *State) IsLeaseActive(staleAfter time.Duration) bool {
	if s == nil || s.ID == "" {
		return false
	}

	return time.Since(s.Started) < staleAfter
}
