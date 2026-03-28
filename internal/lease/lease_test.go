package lease_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestAcquire(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l, err := lease.Read(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l.ID != "span-abc" {
		t.Errorf("ID = %q, want %q", l.ID, "span-abc")
	}

	if l.ReplayCount() != 0 {
		t.Errorf("ReplayCount = %d, want %d", l.ReplayCount(), 0)
	}
}

func TestAcquire_WithReplayCount(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l, err := lease.Read(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l.ReplayCount() != 3 {
		t.Errorf("ReplayCount = %d, want %d", l.ReplayCount(), 3)
	}
}

func TestAcquire_NotFound(t *testing.T) {
	store := tests.NewMockStore[string]()

	if err := lease.Acquire(t.Context(), store, "nonexistent", "span-abc", 0); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestAcquire_UpdateError(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))
	store.UpdateErr = fmt.Errorf("update denied")

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestRead_NoAnnotations(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	state, err := lease.Read(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state != nil {
		t.Errorf("expected nil state, got %+v", state)
	}
}

func TestRead_NotFound(t *testing.T) {
	store := tests.NewMockStore[string]()

	if _, err := lease.Read(t.Context(), store, "nonexistent"); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestRead_MalformedTimestamp(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](map[string]string{
		lease.IDAnnotation:      "span-abc",
		lease.StartedAnnotation: "not-a-timestamp",
	}))

	if _, err := lease.Read(t.Context(), store, "wt-1"); err == nil {
		t.Fatal("expected error for malformed timestamp")
	}
}

func TestRelease(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	if err := lease.Release(t.Context(), store, "wt-1"); err != nil {
		t.Fatalf("releasing: %v", err)
	}

	state, err := lease.Read(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state != nil {
		t.Errorf("expected nil state after release, got %+v", state)
	}
}

func TestRelease_NotFound(t *testing.T) {
	store := tests.NewMockStore[string]()

	if err := lease.Release(t.Context(), store, "nonexistent"); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestRelease_UpdateError(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))
	store.UpdateErr = fmt.Errorf("update denied")

	if err := lease.Release(t.Context(), store, "wt-1"); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestHeldBy(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	holds, err := lease.HeldBy(t.Context(), store, "wt-1", "span-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !holds {
		t.Error("expected to hold lease")
	}
}

func TestHeldBy_DifferentID(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	holds, err := lease.HeldBy(t.Context(), store, "wt-1", "span-other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if holds {
		t.Error("expected not to hold lease")
	}
}

func TestHeldBy_ReadError(t *testing.T) {
	store := tests.NewMockStore[string]()

	if _, err := lease.HeldBy(t.Context(), store, "nonexistent", "span-abc"); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestIsLeaseActive(t *testing.T) {
	l := &lease.State{ID: "span-abc", Started: time.Now()}

	if !l.IsLeaseActive(10 * time.Minute) {
		t.Error("expected active lease")
	}
}

func TestIsLeaseActive_Stale(t *testing.T) {
	l := &lease.State{ID: "span-abc", Started: time.Now().Add(-20 * time.Minute)}

	if l.IsLeaseActive(10 * time.Minute) {
		t.Error("expected stale lease")
	}
}

func TestIsLeaseActive_Empty(t *testing.T) {
	l := &lease.State{}

	if l.IsLeaseActive(10 * time.Minute) {
		t.Error("expected empty lease to be inactive")
	}
}

func TestIsLeaseActive_Nil(t *testing.T) {
	var l *lease.State

	if l.IsLeaseActive(10 * time.Minute) {
		t.Error("expected nil lease to be inactive")
	}
}

func TestReplayCount_Nil(t *testing.T) {
	var l *lease.State

	if l.ReplayCount() != 0 {
		t.Errorf("expected 0, got %d", l.ReplayCount())
	}
}
