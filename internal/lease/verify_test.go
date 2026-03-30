//revive:disable:package-comments
package lease_test

import (
	"testing"

	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestVerifyFunc(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	newStore := func(_ string) object.Store[string] { return store }
	verify := lease.NewVerifyFunc(newStore)

	holds, err := verify(t.Context(), "tenant-a", "wt-1", "span-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !holds {
		t.Error("expected to hold lease")
	}
}

func TestVerifyFunc_DifferentID(t *testing.T) {
	store := tests.NewMockStore[string]()
	store.Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), store, "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	newStore := func(_ string) object.Store[string] { return store }
	verify := lease.NewVerifyFunc(newStore)

	holds, err := verify(t.Context(), "tenant-a", "wt-1", "span-other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if holds {
		t.Error("expected not to hold lease")
	}
}

func TestVerifyFunc_Error(t *testing.T) {
	store := tests.NewMockStore[string]()
	newStore := func(_ string) object.Store[string] { return store }
	verify := lease.NewVerifyFunc(newStore)

	if _, err := verify(t.Context(), "tenant-a", "nonexistent", "span-abc"); err == nil {
		t.Fatal("expected error for missing object")
	}
}

func TestVerifyFunc_PassesNamespace(t *testing.T) {
	stores := map[string]*tests.MockStore[string]{
		"tenant-a": tests.NewMockStore[string](),
		"tenant-b": tests.NewMockStore[string](),
	}
	stores["tenant-a"].Add("wt-1", tests.NewMockObject[string](nil))

	if err := lease.Acquire(t.Context(), stores["tenant-a"], "wt-1", "span-abc", 0); err != nil {
		t.Fatalf("acquiring: %v", err)
	}

	var calledNS string
	newStore := func(ns string) object.Store[string] {
		calledNS = ns
		return stores[ns]
	}
	verify := lease.NewVerifyFunc(newStore)

	holds, err := verify(t.Context(), "tenant-a", "wt-1", "span-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !holds {
		t.Error("expected to hold lease")
	}

	if calledNS != "tenant-a" {
		t.Errorf("namespace = %q, want %q", calledNS, "tenant-a")
	}
}
