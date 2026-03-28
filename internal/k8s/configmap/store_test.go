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
	"github.com/katastroma/phortizo/internal/object"
)

type notAConfigMap struct{}

func (n *notAConfigMap) GetName() string                    { return "" }
func (n *notAConfigMap) SetName(_ string)                   {}
func (n *notAConfigMap) GetAnnotations() map[string]string  { return nil }
func (n *notAConfigMap) SetAnnotations(_ map[string]string) {}
func (n *notAConfigMap) GetLabels() map[string]string       { return nil }
func (n *notAConfigMap) SetLabels(_ map[string]string)      {}
func (n *notAConfigMap) GetData() map[string]string         { return nil }
func (n *notAConfigMap) SetData(_ map[string]string)        {}

var _ object.Resource = (*notAConfigMap)(nil)

func testConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "cm-1", Namespace: "tenant-a"},
	}
}

func TestStore_Get(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	store := configmap.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "cm-1")
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

	obj, err := store.Get(t.Context(), "cm-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.SetAnnotations(map[string]string{"test": "value"})
	if err = store.Update(t.Context(), obj); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := store.Get(t.Context(), "cm-1")
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

	obj, err := store.Get(t.Context(), "cm-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if err = store.Update(t.Context(), obj); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestStore_Put_Create(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	r := configmap.NewResource("my-config")
	r.SetData(map[string]string{"key": "value"})
	r.SetLabels(map[string]string{"type": "test"})

	if err := store.Put(t.Context(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, err := store.Get(t.Context(), "my-config")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if obj.GetData()["key"] != "value" {
		t.Errorf("key = %q, want %q", obj.GetData()["key"], "value")
	}
}

func TestStore_Put_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	store := configmap.NewStore(k8s, "tenant-a")

	r := configmap.NewResource("cm-1")
	r.SetData(map[string]string{"key": "updated"})

	if err := store.Put(t.Context(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, err := store.Get(t.Context(), "cm-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if obj.GetData()["key"] != "updated" {
		t.Errorf("key = %q, want %q", obj.GetData()["key"], "updated")
	}
}

func TestStore_Put_WrongType(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	if err := store.Put(t.Context(), &notAConfigMap{}); err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestStore_Put_CreateError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor(
		"create", "configmaps",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("create denied")
		},
	)
	store := configmap.NewStore(k8s, "tenant-a")

	r := configmap.NewResource("my-config")
	if err := store.Put(t.Context(), r); err == nil {
		t.Fatal("expected error from failing create")
	}
}

func TestStore_Put_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(testConfigMap())
	k8s.PrependReactor(
		"update", "configmaps",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("update denied")
		},
	)
	store := configmap.NewStore(k8s, "tenant-a")

	r := configmap.NewResource("cm-1")
	if err := store.Put(t.Context(), r); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestStore_New(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	r := store.New("my-resource")
	if r.GetName() != "my-resource" {
		t.Errorf("Name = %q, want %q", r.GetName(), "my-resource")
	}
}

func labeledConfigMap(name, namespace string, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func TestStore_List(t *testing.T) {
	typeLabel := map[string]string{"type": "target"}
	k8s := fake.NewSimpleClientset(
		labeledConfigMap("a", "tenant-a", typeLabel),
		labeledConfigMap("b", "tenant-a", typeLabel),
		labeledConfigMap("c", "tenant-a", map[string]string{"type": "other"}),
	)
	store := configmap.NewStore(k8s, "tenant-a")

	results, err := store.List(t.Context(), typeLabel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestStore_List_Empty(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := configmap.NewStore(k8s, "tenant-a")

	results, err := store.List(t.Context(), map[string]string{"type": "target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestStore_List_Error(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor(
		"list", "configmaps",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("list denied")
		},
	)
	store := configmap.NewStore(k8s, "tenant-a")

	_, err := store.List(t.Context(), nil)
	if err == nil {
		t.Fatal("expected error from failing list")
	}
}
