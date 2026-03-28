package provider_test

import (
	"bytes"
	"testing"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/katastroma/phortizo/internal/credential/provider"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestSSHKey_Authenticate(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	cred := provider.NewSSHKey(pemBytes)

	auth, err := cred.Authenticate(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := auth.(*gitssh.PublicKeys); !ok {
		t.Fatalf("got %T, want *gitssh.PublicKeys", auth)
	}
}

func TestSSHKey_Authenticate_InvalidKey(t *testing.T) {
	cred := provider.NewSSHKey([]byte("not-a-valid-key"))

	_, err := cred.Authenticate(t.Context(), nil)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestSSHKey_MarshalSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	cred := provider.NewSSHKey(pemBytes)
	data := cred.MarshalSecret()

	if string(data["type"]) != provider.TypeSSHKey {
		t.Errorf("type = %q, want %q", data["type"], provider.TypeSSHKey)
	}

	if !bytes.Equal(data["private-key"], pemBytes) {
		t.Error("private-key mismatch")
	}
}

func TestSSHKey_FromSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	original := provider.NewSSHKey(pemBytes)
	data := original.MarshalSecret()

	cred, err := provider.SSHKeyFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restored, ok := cred.(*provider.SSHKey)
	if !ok {
		t.Fatalf("expected *provider.SSHKey, got %T", cred)
	}

	restoredData := restored.MarshalSecret()
	if !bytes.Equal(restoredData["private-key"], data["private-key"]) {
		t.Error("round-trip mismatch")
	}
}

func TestSSHKey_FromSecret_MissingKey(t *testing.T) {
	data := map[string][]byte{"type": []byte(provider.TypeSSHKey)}

	_, err := provider.SSHKeyFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing private-key")
	}
}
