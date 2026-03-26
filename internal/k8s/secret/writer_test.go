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
)

func TestPut_Create(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	writer := secret.NewWriter(k8s)

	data := map[string][]byte{"token": []byte("ghp_abc123")}
	err := writer.Put(t.Context(), "tenant-a", "repo-cred", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := k8s.CoreV1().Secrets("tenant-a").Get(t.Context(), "repo-cred", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created: %v", err)
	}

	if string(s.Data["token"]) != "ghp_abc123" {
		t.Errorf("token = %q, want %q", s.Data["token"], "ghp_abc123")
	}
}

func TestPut_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data:       map[string][]byte{"token": []byte("old-token")},
	})
	writer := secret.NewWriter(k8s)

	data := map[string][]byte{"token": []byte("new-token")}
	err := writer.Put(t.Context(), "tenant-a", "repo-cred", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := k8s.CoreV1().Secrets("tenant-a").Get(t.Context(), "repo-cred", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("reading secret: %v", err)
	}

	if string(s.Data["token"]) != "new-token" {
		t.Errorf("token = %q, want %q", s.Data["token"], "new-token")
	}
}

func TestPut_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data:       map[string][]byte{"token": []byte("old-token")},
	})
	k8s.PrependReactor("update", "secrets", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("update denied")
	})

	writer := secret.NewWriter(k8s)
	err := writer.Put(t.Context(), "tenant-a", "repo-cred", map[string][]byte{"token": []byte("new")})
	if err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestPut_CreateError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor("create", "secrets", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("create denied")
	})

	writer := secret.NewWriter(k8s)
	err := writer.Put(t.Context(), "tenant-a", "repo-cred", map[string][]byte{"token": []byte("ghp_abc123")})
	if err == nil {
		t.Fatal("expected error from failing create")
	}
}
