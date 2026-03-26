package credential_test

import (
	"bytes"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/github"
	credential "github.com/katastroma/phortizo/internal/onboarding/repo_credential"
)

func TestGitHubToken_Authenticate(t *testing.T) {
	cred := credential.NewGitHubToken("ghp_abc123")

	auth, err := cred.Authenticate(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	basic, ok := auth.(*githttp.BasicAuth)
	if !ok {
		t.Fatalf("got %T, want *githttp.BasicAuth", auth)
	}

	if basic.Username != github.TokenUserName {
		t.Errorf("Username = %q, want %q", basic.Username, github.TokenUserName)
	}

	if basic.Password != "ghp_abc123" {
		t.Errorf("Password = %q, want %q", basic.Password, "ghp_abc123")
	}
}

func TestGitHubToken_MarshalSecret(t *testing.T) {
	cred := credential.NewGitHubToken("ghp_abc123")
	data := cred.MarshalSecret()

	if string(data["type"]) != credential.TypeGitHubToken {
		t.Errorf("type = %q, want %q", data["type"], credential.TypeGitHubToken)
	}

	if string(data["token"]) != "ghp_abc123" {
		t.Errorf("token = %q, want %q", data["token"], "ghp_abc123")
	}
}

func TestGitHubToken_FromSecret(t *testing.T) {
	original := credential.NewGitHubToken("ghp_abc123")
	data := original.MarshalSecret()

	restored, err := credential.GitHubTokenFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restoredData := restored.MarshalSecret()
	if !bytes.Equal(restoredData["token"], data["token"]) {
		t.Errorf("round-trip mismatch: got %q, want %q", restoredData["token"], data["token"])
	}
}

func TestGitHubToken_FromSecret_MissingKey(t *testing.T) {
	data := map[string][]byte{"type": []byte(credential.TypeGitHubToken)}

	_, err := credential.GitHubTokenFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing token key")
	}
}
