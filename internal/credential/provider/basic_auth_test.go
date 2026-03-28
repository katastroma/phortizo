package provider_test

import (
	"bytes"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/credential/provider"
)

func TestBasicAuth_Authenticate(t *testing.T) {
	cred := provider.NewBasicAuth("user", "pass")

	auth, err := cred.Authenticate(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	basic, ok := auth.(*githttp.BasicAuth)
	if !ok {
		t.Fatalf("got %T, want *githttp.BasicAuth", auth)
	}

	if basic.Username != "user" {
		t.Errorf("Username = %q, want %q", basic.Username, "user")
	}

	if basic.Password != "pass" {
		t.Errorf("Password = %q, want %q", basic.Password, "pass")
	}
}

func TestBasicAuth_MarshalSecret(t *testing.T) {
	cred := provider.NewBasicAuth("user", "pass")
	data := cred.MarshalSecret()

	if string(data["type"]) != provider.TypeBasicAuth {
		t.Errorf("type = %q, want %q", data["type"], provider.TypeBasicAuth)
	}

	if string(data["username"]) != "user" {
		t.Errorf("username = %q, want %q", data["username"], "user")
	}

	if string(data["password"]) != "pass" {
		t.Errorf("password = %q, want %q", data["password"], "pass")
	}
}

func TestBasicAuth_FromSecret(t *testing.T) {
	original := provider.NewBasicAuth("user", "pass")
	data := original.MarshalSecret()

	cred, err := provider.BasicAuthFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restored, ok := cred.(*provider.BasicAuth)
	if !ok {
		t.Fatalf("expected *provider.BasicAuth, got %T", cred)
	}

	restoredData := restored.MarshalSecret()
	if !bytes.Equal(restoredData["username"], data["username"]) {
		t.Errorf("username round-trip mismatch: got %q, want %q", restoredData["username"], data["username"])
	}

	if !bytes.Equal(restoredData["password"], data["password"]) {
		t.Errorf("password round-trip mismatch: got %q, want %q", restoredData["password"], data["password"])
	}
}

func TestBasicAuth_FromSecret_MissingUsername(t *testing.T) {
	data := map[string][]byte{
		"type":     []byte(provider.TypeBasicAuth),
		"password": []byte("pass"),
	}

	_, err := provider.BasicAuthFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing username key")
	}
}

func TestBasicAuth_FromSecret_MissingPassword(t *testing.T) {
	data := map[string][]byte{
		"type":     []byte(provider.TypeBasicAuth),
		"username": []byte("user"),
	}

	_, err := provider.BasicAuthFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing password key")
	}
}
