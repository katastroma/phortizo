package configmap_test

import (
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/source"
)

func bareConfigMap(name, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{configmap.TypeLabel: source.TypeLabel},
		},
		Data: map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		},
	}
}

func TestAcquireLease(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lease, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.WatchTargetLeaseID != "span-abc" {
		t.Errorf("WatchTargetLeaseID = %q, want %q", lease.WatchTargetLeaseID, "span-abc")
	}

	if lease.ReplayCount != 0 {
		t.Errorf("ReplayCount = %d, want %d", lease.ReplayCount, 0)
	}
}

func TestAcquireLease_WithReplayCount(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lease, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.ReplayCount != 3 {
		t.Errorf("ReplayCount = %d, want %d", lease.ReplayCount, 3)
	}
}

func TestAcquireLease_ConfigMapNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "nonexistent", "span-abc", 0)
	if err == nil {
		t.Fatal("expected error for missing configmap")
	}
}

func TestReadLease_NoAnnotations(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	lease, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.WatchTargetLeaseID != "" {
		t.Errorf("WatchTargetLeaseID = %q, want empty", lease.WatchTargetLeaseID)
	}
}

func TestReadLease_ConfigMapNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	_, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing configmap")
	}
}

func TestReleaseLease(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 0)
	if err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	err = configmap.ReleaseLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("releasing: %v", err)
	}

	lease, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("reading: %v", err)
	}

	if lease.WatchTargetLeaseID != "" {
		t.Errorf("WatchTargetLeaseID = %q, want empty after release", lease.WatchTargetLeaseID)
	}
}

func TestReleaseLease_ConfigMapNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	err := configmap.ReleaseLease(t.Context(), k8s, "tenant-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing configmap")
	}
}

func TestHoldsLease(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 0)
	if err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	holds, err := configmap.HoldsLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !holds {
		t.Error("expected to hold lease")
	}
}

func TestHoldsLease_DifferentLeaseID(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 0)
	if err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	holds, err := configmap.HoldsLease(t.Context(), k8s, "tenant-a", "wt-1", "span-other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if holds {
		t.Error("expected not to hold lease")
	}
}

func TestHoldsLease_ReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	_, err := configmap.HoldsLease(t.Context(), k8s, "tenant-a", "nonexistent", "span-abc")
	if err == nil {
		t.Fatal("expected error for missing configmap")
	}
}

func TestIsActive(t *testing.T) {
	lease := configmap.Lease{
		WatchTargetLeaseID: "span-abc",
		Started:            time.Now(),
	}

	if !lease.IsLeaseActive(10 * time.Minute) {
		t.Error("expected active lease")
	}
}

func TestIsActive_Stale(t *testing.T) {
	lease := configmap.Lease{
		WatchTargetLeaseID: "span-abc",
		Started:            time.Now().Add(-20 * time.Minute),
	}

	if lease.IsLeaseActive(10 * time.Minute) {
		t.Error("expected stale lease")
	}
}

func TestIsActive_Empty(t *testing.T) {
	lease := configmap.Lease{}

	if lease.IsLeaseActive(10 * time.Minute) {
		t.Error("expected empty lease to be inactive")
	}
}

func TestReadLease_MalformedTimestamp(t *testing.T) {
	cm := bareConfigMap("wt-1", "tenant-a")
	cm.Annotations = map[string]string{
		configmap.WatchTargetLeaseIDAnnotation:      "span-abc",
		configmap.WatchTargetLeaseStartedAnnotation: "not-a-timestamp",
	}
	k8s := fake.NewSimpleClientset(cm)

	lease, err := configmap.ReadLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.WatchTargetLeaseID != "" {
		t.Errorf("expected empty lease for malformed timestamp, got WatchTargetLeaseID = %q", lease.WatchTargetLeaseID)
	}
}

func TestAcquireLease_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))
	k8s.PrependReactor("update", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("update denied")
	})

	err := configmap.AcquireLease(t.Context(), k8s, "tenant-a", "wt-1", "span-abc", 0)
	if err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestReleaseLease_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(bareConfigMap("wt-1", "tenant-a"))
	k8s.PrependReactor("update", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("update denied")
	})

	err := configmap.ReleaseLease(t.Context(), k8s, "tenant-a", "wt-1")
	if err == nil {
		t.Fatal("expected error from failing update")
	}
}
