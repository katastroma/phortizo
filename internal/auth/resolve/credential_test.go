package resolve

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/katastroma/phortizo/internal/vault"
)

func testSSHPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

func TestFromCredential_AppInstallation(t *testing.T) {
	exchanger := &stubExchanger{token: "app-token"}
	cred := &vault.Credential{Type: vault.AppInstallation, InstallationID: 123}

	method, err := FromCredential(t.Context(), cred, exchanger, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method == nil {
		t.Fatal("expected non-nil method")
	}
}

func TestFromCredential_AppInstallation_ExchangeError(t *testing.T) {
	exchanger := &stubExchanger{err: errExchange}
	cred := &vault.Credential{Type: vault.AppInstallation, InstallationID: 123}

	if _, err := FromCredential(t.Context(), cred, exchanger, nil); err == nil {
		t.Fatal("expected error for failed exchange")
	}
}

func TestFromCredential_TenantApp(t *testing.T) {
	cred := &vault.Credential{
		Type:           vault.TenantApp,
		ClientID:       "client-id",
		PrivateKeyPEM:  []byte("unused"),
		InstallationID: 123,
	}

	method, err := FromCredential(t.Context(), cred, nil, &stubFactory{token: "tenant-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method == nil {
		t.Fatal("expected non-nil method")
	}
}

func TestFromCredential_TenantApp_ExchangeError(t *testing.T) {
	cred := &vault.Credential{
		Type:           vault.TenantApp,
		ClientID:       "client-id",
		PrivateKeyPEM:  []byte("unused"),
		InstallationID: 123,
	}

	if _, err := FromCredential(t.Context(), cred, nil, &stubFactory{err: errExchange}); err == nil {
		t.Fatal("expected error for failed tenant app exchange")
	}
}

func TestFromCredential_Token(t *testing.T) {
	cred := &vault.Credential{Type: vault.Token, Token: "ghp_abc"}

	method, err := FromCredential(t.Context(), cred, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method == nil {
		t.Fatal("expected non-nil method")
	}
}

func TestFromCredential_BasicAuth(t *testing.T) {
	cred := &vault.Credential{Type: vault.BasicAuth, Username: "user", Password: "pass"}

	method, err := FromCredential(t.Context(), cred, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method == nil {
		t.Fatal("expected non-nil method")
	}
}

func TestFromCredential_SSH(t *testing.T) {
	cred := &vault.Credential{Type: vault.SSH, SSHKeyPEM: testSSHPEM(t)}

	method, err := FromCredential(t.Context(), cred, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method == nil {
		t.Fatal("expected non-nil method")
	}
}

func TestFromCredential_SSH_InvalidKey(t *testing.T) {
	cred := &vault.Credential{Type: vault.SSH, SSHKeyPEM: []byte("bad key")}

	if _, err := FromCredential(t.Context(), cred, nil, nil); err == nil {
		t.Fatal("expected error for invalid SSH key")
	}
}

func TestFromCredential_UnsupportedType(t *testing.T) {
	cred := &vault.Credential{Type: "unknown"}

	if _, err := FromCredential(t.Context(), cred, nil, nil); err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
