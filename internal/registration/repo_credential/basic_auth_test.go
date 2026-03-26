package credential_test

import (
	"bytes"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	credential "github.com/katastroma/phortizo/internal/registration/repo_credential"
)

func TestBasicAuth_Authenticate(t *testing.T) {
	cred := credential.NewBasicAuth("user", "pass")

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
	cred := credential.NewBasicAuth("user", "pass")
	data := cred.MarshalSecret()

	if string(data["type"]) != credential.TypeBasicAuth {
		t.Errorf("type = %q, want %q", data["type"], credential.TypeBasicAuth)
	}

	if string(data["username"]) != "user" {
		t.Errorf("username = %q, want %q", data["username"], "user")
	}

	if string(data["password"]) != "pass" {
		t.Errorf("password = %q, want %q", data["password"], "pass")
	}
}

func TestBasicAuth_FromSecret(t *testing.T) {
	original := credential.NewBasicAuth("user", "pass")
	data := original.MarshalSecret()

	restored, err := credential.BasicAuthFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
		"type":     []byte(credential.TypeBasicAuth),
		"password": []byte("pass"),
	}

	_, err := credential.BasicAuthFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing username key")
	}
}

func TestBasicAuth_FromSecret_MissingPassword(t *testing.T) {
	data := map[string][]byte{
		"type":     []byte(credential.TypeBasicAuth),
		"username": []byte("user"),
	}

	_, err := credential.BasicAuthFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing password key")
	}
}
