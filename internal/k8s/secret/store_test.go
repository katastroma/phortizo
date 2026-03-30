//revive:disable:package-comments
package secret_test

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/object"
)

type notASecret struct{}

func (n *notASecret) GetName() string                    { return "" }
func (n *notASecret) SetName(_ string)                   {}
func (n *notASecret) GetAnnotations() map[string]string  { return nil }
func (n *notASecret) SetAnnotations(_ map[string]string) {}
func (n *notASecret) GetLabels() map[string]string       { return nil }
func (n *notASecret) SetLabels(_ map[string]string)      {}
func (n *notASecret) GetData() map[string][]byte         { return nil }
func (n *notASecret) SetData(_ map[string][]byte)        {}

var _ object.Resource[[]byte] = (*notASecret)(nil)

func testSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sec-1", Namespace: "tenant-a"},
	}
}

func TestStore_Get(t *testing.T) {
	k8s := fake.NewSimpleClientset(testSecret())
	store := secret.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "sec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{"test": "value"})
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := secret.NewStore(k8s, "tenant-a")

	if _, err := store.Get(t.Context(), "nonexistent"); err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestStore_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(testSecret())
	store := secret.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "sec-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.SetAnnotations(map[string]string{"test": "value"})
	if err = store.Update(t.Context(), obj); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := store.Get(t.Context(), "sec-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if updated.GetAnnotations()["test"] != "value" {
		t.Errorf("annotation = %q, want %q", updated.GetAnnotations()["test"], "value")
	}
}

func TestStore_Update_WrongType(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := secret.NewStore(k8s, "tenant-a")

	if err := store.Update(t.Context(), &notASecret{}); err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestStore_Update_Error(t *testing.T) {
	k8s := fake.NewSimpleClientset(testSecret())
	k8s.PrependReactor(
		"update", "secrets",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("update denied")
		},
	)
	store := secret.NewStore(k8s, "tenant-a")

	obj, err := store.Get(t.Context(), "sec-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if err = store.Update(t.Context(), obj); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestStore_Put_Create(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := secret.NewStore(k8s, "tenant-a")

	r := secret.NewResource("my-secret")
	r.SetData(map[string][]byte{"token": []byte("ghp_abc123")})
	r.SetLabels(map[string]string{"type": "test"})

	if err := store.Put(t.Context(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, err := store.Get(t.Context(), "my-secret")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if string(obj.GetData()["token"]) != "ghp_abc123" {
		t.Errorf("token = %q, want %q", obj.GetData()["token"], "ghp_abc123")
	}
}

func TestStore_Put_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(testSecret())
	store := secret.NewStore(k8s, "tenant-a")

	r := secret.NewResource("sec-1")
	r.SetData(map[string][]byte{"token": []byte("updated")})

	if err := store.Put(t.Context(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, err := store.Get(t.Context(), "sec-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if string(obj.GetData()["token"]) != "updated" {
		t.Errorf("token = %q, want %q", obj.GetData()["token"], "updated")
	}
}

func TestStore_Put_WrongType(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := secret.NewStore(k8s, "tenant-a")

	if err := store.Put(t.Context(), &notASecret{}); err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestStore_Put_CreateError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor(
		"create", "secrets",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("create denied")
		},
	)
	store := secret.NewStore(k8s, "tenant-a")

	r := secret.NewResource("my-secret")
	if err := store.Put(t.Context(), r); err == nil {
		t.Fatal("expected error from failing create")
	}
}

func TestStore_Put_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(testSecret())
	k8s.PrependReactor(
		"update", "secrets",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("update denied")
		},
	)
	store := secret.NewStore(k8s, "tenant-a")

	r := secret.NewResource("sec-1")
	if err := store.Put(t.Context(), r); err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestStore_New(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	store := secret.NewStore(k8s, "tenant-a")

	r := store.New("my-resource")
	if r.GetName() != "my-resource" {
		t.Errorf("Name = %q, want %q", r.GetName(), "my-resource")
	}
}

func labeledSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func TestStore_List(t *testing.T) {
	typeLabel := map[string]string{"type": "credential"}
	k8s := fake.NewSimpleClientset(
		labeledSecret("a", "tenant-a", typeLabel),
		labeledSecret("b", "tenant-a", typeLabel),
		labeledSecret("c", "tenant-a", map[string]string{"type": "other"}),
	)
	store := secret.NewStore(k8s, "tenant-a")

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
	store := secret.NewStore(k8s, "tenant-a")

	results, err := store.List(t.Context(), map[string]string{"type": "credential"})
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
		"list", "secrets",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("list denied")
		},
	)
	store := secret.NewStore(k8s, "tenant-a")

	_, err := store.List(t.Context(), nil)
	if err == nil {
		t.Fatal("expected error from failing list")
	}
}
