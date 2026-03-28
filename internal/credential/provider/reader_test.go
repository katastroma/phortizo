package provider_test

import (
	"testing"

	"github.com/katastroma/phortizo/internal/credential/provider"
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
		"type":  []byte(provider.TypeGitHubToken),
		"token": []byte("ghp_abc123"),
	})

	reader := provider.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*provider.GitHubToken); !ok {
		t.Fatalf("expected *provider.GitHubToken, got %T", cred)
	}
}

func TestGet_BasicAuth(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":     []byte(provider.TypeBasicAuth),
		"username": []byte("user"),
		"password": []byte("pass"),
	})

	reader := provider.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*provider.BasicAuth); !ok {
		t.Fatalf("expected *provider.BasicAuth, got %T", cred)
	}
}

func TestGet_SSHKey(t *testing.T) {
	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":        []byte(provider.TypeSSHKey),
		"private-key": []byte("fake-pem-bytes"),
	})

	reader := provider.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*provider.SSHKey); !ok {
		t.Fatalf("expected *provider.SSHKey, got %T", cred)
	}
}

func TestGet_GitHubAppTenant(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	store := tests.NewMockStore[[]byte]()
	seedSecret(store, "repo-cred", map[string][]byte{
		"type":            []byte(provider.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     pemBytes,
		"installation-id": []byte("12345"),
	})

	reader := provider.NewReader(nil)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*provider.GitHubAppTenant); !ok {
		t.Fatalf("expected *provider.GitHubAppTenant, got %T", cred)
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
		"type":            []byte(provider.TypeGitHubAppPlatform),
		"installation-id": []byte("67890"),
	})

	reader := provider.NewReader(platformApp)
	cred, err := reader.Get(t.Context(), store, "repo-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cred.(*provider.GitHubAppPlatform); !ok {
		t.Fatalf("expected *provider.GitHubAppPlatform, got %T", cred)
	}
}

func TestGet_SecretNotFound(t *testing.T) {
	store := tests.NewMockStore[[]byte]()

	reader := provider.NewReader(nil)
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

	reader := provider.NewReader(nil)
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

	reader := provider.NewReader(nil)
	_, err := reader.Get(t.Context(), store, "repo-cred")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
