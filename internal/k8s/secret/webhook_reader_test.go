package secret_test

import (
	"bytes"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/katastroma/phortizo/internal/k8s/secret"
)

func TestReadWebhookSecret(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "tenant-a"},
		Data:       map[string][]byte{secret.WebhookSecretKey: []byte("hmac-secret-bytes")},
	})

	got, err := secret.ReadWebhookSecret(t.Context(), k8s, "tenant-a", "webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(got, []byte("hmac-secret-bytes")) {
		t.Errorf("got %q, want %q", got, "hmac-secret-bytes")
	}
}

func TestReadWebhookSecret_NotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	_, err := secret.ReadWebhookSecret(t.Context(), k8s, "tenant-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestReadWebhookSecret_MissingKey(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "tenant-a"},
		Data:       map[string][]byte{"wrong-key": []byte("hmac-secret-bytes")},
	})

	_, err := secret.ReadWebhookSecret(t.Context(), k8s, "tenant-a", "webhook")
	if err == nil {
		t.Fatal("expected error for missing webhook-secret key")
	}
}
