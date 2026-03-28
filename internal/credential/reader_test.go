package credential_test

import (
	"testing"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/tests"
)

func seedSecret(store *tests.MockStore[[]byte], name string, data map[string][]byte) {
	obj := tests.NewMockObject[[]byte](nil)
	obj.SetData(data)
	store.Add(name, obj)
}

func TestGet_GitHubToken(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":  []byte(credential.TypeGitHubToken),
		"token": []byte("ghp_abc123"),
	})

	reader := credential.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.GitHubToken); !ok {
		t.Fatalf("expected *credential.GitHubToken, got %T", cred)
	}
}

func TestGet_BasicAuth(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":     []byte(credential.TypeBasicAuth),
		"username": []byte("user"),
		"password": []byte("pass"),
	})

	reader := credential.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.BasicAuth); !ok {
		t.Fatalf("expected *credential.BasicAuth, got %T", cred)
	}
}

func TestGet_SSHKey(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":        []byte(credential.TypeSSHKey),
		"private-key": []byte("fake-pem-bytes"),
	})

	reader := credential.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.SSHKey); !ok {
		t.Fatalf("expected *credential.SSHKey, got %T", cred)
	}
}

func TestGet_GitHubAppTenant(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     pemBytes,
		"installation-id": []byte("12345"),
	})

	reader := credential.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
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

	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppPlatform),
		"installation-id": []byte("67890"),
	})

	reader := credential.NewReader(platformApp)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*credential.GitHubAppPlatform); !ok {
		t.Fatalf("expected *credential.GitHubAppPlatform, got %T", cred)
	}
}

func TestGet_SecretNotFound(t *testing.T) {
	store := tests.NewMockStore[[]byte]()

	reader := credential.NewReader(nil)
	_, err := reader.Get(t.Context(), store, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestGet_MissingType(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"token": []byte("ghp_abc123"),
	})

	reader := credential.NewReader(nil)
	_, err := reader.Get(t.Context(), store, "repo-cred")
	if err == nil {
		t.Fatal("expected error for missing type key")
	}
}

func TestGet_UnknownType(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type": []byte("unknown"),
	})

	reader := credential.NewReader(nil)
	_, err := reader.Get(t.Context(), store, "repo-cred")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
