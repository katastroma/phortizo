// Package lease controls operations for processing resource ownership.
//
// Kubernetes provides a first-class Lease resource (coordination.k8s.io/v1)
// for leader election and distributed locking. We use annotations instead
// because our objects have external lifecycles. Their lifecycles are managed by
// tenant onboarding, not the processor. Storing lease state as annotations
// avoids creating and garbage-collecting a separate Lease resource per source
// target per processing instance. The object's resourceVersion provides
// optimistic concurrency for free.
package lease

import (
	"time"
)

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

// IsLeaseActive checks if a lease is held and not stale.
func (s *State) IsLeaseActive(staleAfter time.Duration) bool {
	if s == nil || s.ID == "" {
		return false
	}

	return time.Since(s.Started) < staleAfter
}
