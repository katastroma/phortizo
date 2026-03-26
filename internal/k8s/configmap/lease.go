// Lease operations for pipeline run ownership.
//
// Kubernetes provides a first-class Lease resource (coordination.k8s.io/v1)
// for leader election and distributed locking. We use ConfigMap annotations
// instead because the watch target ConfigMap already exists for the lifetime
// of the watch target — its lifecycle is managed by tenant registration, not
// the pipeline. Storing lease state as annotations avoids creating and
// garbage-collecting a separate Lease resource per watch target per pipeline
// run. The ConfigMap's resourceVersion provides optimistic concurrency for
// free.

//revive:disable:package-comments
package configmap

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// LeaseRunIDAnnotation is the run ID that currently owns the watch target.
	LeaseRunIDAnnotation = "katastroma.org/active-run-id"

	// LeaseStartedAnnotation is the timestamp when the lease was acquired.
	LeaseStartedAnnotation = "katastroma.org/active-run-started"

	// LeaseReplayCountAnnotation tracks how many times this run ID has been replayed.
	LeaseReplayCountAnnotation = "katastroma.org/active-run-replay-count"
)

// Lease represents the current lease state on a watch target ConfigMap.
type Lease struct {
	RunID       string
	Started     time.Time
	ReplayCount int
}

// AcquireLease attempts to take ownership of a watch target ConfigMap. Uses
// optimistic concurrency via resourceVersion to prevent races.
func AcquireLease(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
	runID string,
	replayCount int,
) error {
	configmaps := client.CoreV1().ConfigMaps(namespace)

	cm, err := configmaps.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading configmap %s/%s: %w", namespace, name, err)
	}

	if cm.Annotations == nil {
		cm.Annotations = make(map[string]string)
	}

	cm.Annotations[LeaseRunIDAnnotation] = runID
	cm.Annotations[LeaseStartedAnnotation] = time.Now().UTC().Format(time.RFC3339)
	cm.Annotations[LeaseReplayCountAnnotation] = strconv.Itoa(replayCount)

	_, err = configmaps.Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("acquiring lease on %s/%s: %w", namespace, name, err)
	}

	return nil
}

// ReadLease reads the current lease state from a watch target ConfigMap.
// Returns a zero Lease if no lease annotations are present.
func ReadLease(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
) (Lease, error) {
	configmaps := client.CoreV1().ConfigMaps(namespace)

	cm, err := configmaps.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return Lease{}, fmt.Errorf("reading configmap %s/%s: %w", namespace, name, err)
	}

	return parseLease(cm.Annotations), nil
}

// ReleaseLease clears the lease annotations on a watch target ConfigMap.
func ReleaseLease(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
) error {
	configmaps := client.CoreV1().ConfigMaps(namespace)

	cm, err := configmaps.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("reading configmap %s/%s: %w", namespace, name, err)
	}

	delete(cm.Annotations, LeaseRunIDAnnotation)
	delete(cm.Annotations, LeaseStartedAnnotation)
	delete(cm.Annotations, LeaseReplayCountAnnotation)

	_, err = configmaps.Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("releasing lease on %s/%s: %w", namespace, name, err)
	}

	return nil
}

// HoldsLease checks if the given run ID currently holds the lease.
func HoldsLease(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
	runID string,
) (bool, error) {
	lease, err := ReadLease(ctx, client, namespace, name)
	if err != nil {
		return false, err
	}

	return lease.RunID == runID, nil
}

// IsLeaseActive checks if a lease is held and not stale.
func (l Lease) IsLeaseActive(staleAfter time.Duration) bool {
	if l.RunID == "" {
		return false
	}

	return time.Since(l.Started) < staleAfter
}

func parseLease(annotations map[string]string) Lease {
	runID := annotations[LeaseRunIDAnnotation]
	if runID == "" {
		return Lease{}
	}

	started, err := time.Parse(time.RFC3339, annotations[LeaseStartedAnnotation])
	if err != nil {
		return Lease{}
	}

	replayCount, _ := strconv.Atoi(annotations[LeaseReplayCountAnnotation])

	return Lease{
		RunID:       runID,
		Started:     started,
		ReplayCount: replayCount,
	}
}
