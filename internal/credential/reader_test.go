package credential_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestGet_GitHubToken(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data: map[string][]byte{
			"type":  []byte(credential.TypeGitHubToken),
			"token": []byte("ghp_abc123"),
		},
	})

	reader := credential.NewReader(k8s, nil)
	cred, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.GitHubToken); !ok {
		t.Fatalf("expected *credential.GitHubToken, got %T", cred)
	}
}

func TestGet_BasicAuth(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data: map[string][]byte{
			"type":     []byte(credential.TypeBasicAuth),
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	})

	reader := credential.NewReader(k8s, nil)
	cred, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.BasicAuth); !ok {
		t.Fatalf("expected *credential.BasicAuth, got %T", cred)
	}
}

func TestGet_SSHKey(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data: map[string][]byte{
			"type":        []byte(credential.TypeSSHKey),
			"private-key": []byte("fake-pem-bytes"),
		},
	})

	reader := credential.NewReader(k8s, nil)
	cred, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.SSHKey); !ok {
		t.Fatalf("expected *credential.SSHKey, got %T", cred)
	}
}

func TestGet_GitHubAppTenant(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data: map[string][]byte{
			"type":            []byte(credential.TypeGitHubAppTenant),
			"client-id":       []byte("Iv1.abc123"),
			"private-key":     pemBytes,
			"installation-id": []byte("12345"),
		},
	})

	reader := credential.NewReader(k8s, nil)
	cred, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.GitHubAppTenant); !ok {
		t.Fatalf("expected *credential.GitHubAppTenant, got %T", cred)
	}
}

func TestGet_GitHubAppPlatform(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data: map[string][]byte{
			"type":            []byte(credential.TypeGitHubAppPlatform),
			"installation-id": []byte("67890"),
		},
	})

	reader := credential.NewReader(k8s, platformApp)
	cred, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.GitHubAppPlatform); !ok {
		t.Fatalf("expected *credential.GitHubAppPlatform, got %T", cred)
	}
}

func TestGet_SecretNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	reader := credential.NewReader(k8s, nil)
	_, err := reader.Get(t.Context(), "tenant-a", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestGet_MissingType(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data:       map[string][]byte{"token": []byte("ghp_abc123")},
	})

	reader := credential.NewReader(k8s, nil)
	_, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err == nil {
		t.Fatal("expected error for missing type key")
	}
}

func TestGet_UnknownType(t *testing.T) {
	k8s := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-cred", Namespace: "tenant-a"},
		Data:       map[string][]byte{"type": []byte("unknown")},
	})

	reader := credential.NewReader(k8s, nil)
	_, err := reader.Get(t.Context(), "tenant-a", "repo-cred")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
