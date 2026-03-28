package secret_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/katastroma/phortizo/internal/k8s/secret"
)

func TestRead(t *testing.T) {
	testKeyA := "key-a"
	testKeyB := "key-b"
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "tenant-a"},
		Data:       map[string][]byte{testKeyA: []byte("value-a"), testKeyB: []byte("value-b")},
	})

	got, err := secret.Read(t.Context(), k8s, "tenant-a", "webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(got[testKeyA]) != "value-a" {
		t.Errorf("got %q for key-a, want %q", got[testKeyA], "value-a")
	}

	if string(got[testKeyB]) != "value-b" {
		t.Errorf("got %q for key-b, want %q", got[testKeyB], "value-b")
	}
}

func TestRead_NotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	_, err := secret.Read(t.Context(), k8s, "tenant-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}
