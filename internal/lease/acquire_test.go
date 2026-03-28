//revive:disable:package-comments
package lease_test

import (
	"fmt"
	"testing"

	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestAcquireFunc(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	newStore := func(_ string) object.Store[string] { return store }
	acquire := lease.AcquireFunc(newStore)

	if err := acquire(t.Context(), "tenant-a", "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l, err := lease.Read(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l.ID != "span-abc" {
		t.Errorf("ID = %q, want %q", l.ID, "span-abc")
	}
}

func TestAcquireFunc_Error(t *testing.T) {
	store := tests.NewMockStore[string]()
	newStore := func(_ string) object.Store[string] { return store }
	acquire := lease.AcquireFunc(newStore)

	if err := acquire(t.Context(), "tenant-a", "nonexistent", "span-abc", 0); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestAcquireFunc_PassesNamespace(t *testing.T) {
	stores := map[string]*tests.MockStore[string]{
		"tenant-a": tests.NewMockStore[string](),
		"tenant-b": tests.NewMockStore[string](),
	}
	stores["tenant-a"].Add("wt-1", tests.NewMockObject[string](nil))
	stores["tenant-b"].Add("wt-1", tests.NewMockObject[string](nil))
	stores["tenant-b"].UpdateErr = fmt.Errorf("should not be called")

	newStore := func(ns string) object.Store[string] { return stores[ns] }
	acquire := lease.AcquireFunc(newStore)

	if err := acquire(t.Context(), "tenant-a", "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
