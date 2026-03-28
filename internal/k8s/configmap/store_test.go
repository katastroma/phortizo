//revive:disable:package-comments
package configmap_test

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/lease"
)

type notAConfigMap struct{}

func (n *notAConfigMap) GetAnnotations() map[string]string  { return nil }
func (n *notAConfigMap) SetAnnotations(_ map[string]string) {}

var _ lease.Annotatable = (*notAConfigMap)(nil)

func testConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "wt-1", Namespace: "tenant-a"},
	}
}

func TestStore_Get(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	store := configmap.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{"test": "value"})
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	if _, err := store.Get(t.Context(), "nonexistent"); err == nil {
		t.Fatal("expected error for missing configmap")
	}
}

func TestStore_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	store := configmap.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "wt-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.SetAnnotations(map[string]string{"test": "value"})
	if err = store.Update(t.Context(), obj); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := store.Get(t.Context(), "wt-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if updated.GetAnnotations()["test"] != "value" {
		t.Errorf("annotation = %q, want %q", updated.GetAnnotations()["test"], "value")
	}
}

func TestStore_Update_WrongType(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	bad := &notAConfigMap{}
	if err := store.Update(t.Context(), bad); err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestStore_Update_Error(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	k8s.PrependReactor(
		"update", "configmaps",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("update denied")
		},
	)
	store := configmap.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "wt-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if err = store.Update(t.Context(), obj); err == nil {
		t.Fatal("expected error from failing update")
	}
}
