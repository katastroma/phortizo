package secret_test

import (
	"bytes"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/katastroma/phortizo/internal/k8s/secret"
)

func TestReadSecret(t *testing.T) {
	const testKey = "test-key"
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "tenant-a"},
		Data:       map[string][]byte{testKey: []byte("hmac-secret-bytes")},
	})

	got, err := secret.Read(t.Context(), k8s, "tenant-a", "webhook", testKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(got, []byte("hmac-secret-bytes")) {
		t.Errorf("got %q, want %q", got, "hmac-secret-bytes")
	}
}

func TestReadSecret_NotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	_, err := secret.Read(t.Context(), k8s, "tenant-a", "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestReadSecret_MissingKey(t *testing.T) {
	wrongKey := "test-key"
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "tenant-a"},
		Data:       map[string][]byte{wrongKey: []byte("hmac-secret-bytes")},
	})

	_, err := secret.Read(t.Context(), k8s, "tenant-a", "webhook", "wrong-key")
	if err == nil {
		t.Fatal("expected error for missing webhook-secret key")
	}
}
