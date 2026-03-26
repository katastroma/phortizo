package credential_test

import (
	"bytes"
	"net/http"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/github"
	"github.com/katastroma/phortizo/internal/github/apps"
	credential "github.com/katastroma/phortizo/internal/registration/repo_credential"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestGitHubAppTenant_FromSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     pemBytes,
		"installation-id": []byte("12345"),
	}

	cred, err := credential.GitHubAppTenantFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
}

func TestGitHubAppTenant_FromSecret_MissingClientID(t *testing.T) {
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"private-key":     tests.GenerateRSAPEM(t),
		"installation-id": []byte("12345"),
	}

	_, err := credential.GitHubAppTenantFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing client-id")
	}
}

func TestGitHubAppTenant_FromSecret_MissingPrivateKey(t *testing.T) {
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"installation-id": []byte("12345"),
	}

	_, err := credential.GitHubAppTenantFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing private-key")
	}
}

func TestGitHubAppTenant_FromSecret_MissingInstallationID(t *testing.T) {
	data := map[string][]byte{
		"type":        []byte(credential.TypeGitHubAppTenant),
		"client-id":   []byte("Iv1.abc123"),
		"private-key": tests.GenerateRSAPEM(t),
	}

	_, err := credential.GitHubAppTenantFromSecret(data)
	if err == nil {
		t.Fatal("expected error for missing installation-id")
	}
}

func TestGitHubAppTenant_FromSecret_InvalidInstallationID(t *testing.T) {
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     tests.GenerateRSAPEM(t),
		"installation-id": []byte("not-a-number"),
	}

	_, err := credential.GitHubAppTenantFromSecret(data)
	if err == nil {
		t.Fatal("expected error for invalid installation-id")
	}
}

func TestGitHubAppTenant_FromSecret_InvalidPEM(t *testing.T) {
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     []byte("not-a-pem"),
		"installation-id": []byte("12345"),
	}

	_, err := credential.GitHubAppTenantFromSecret(data)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestGitHubAppTenant_MarshalSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	cred := credential.NewGitHubAppTenant(app, 12345)
	data := cred.MarshalSecret()

	if string(data["type"]) != credential.TypeGitHubAppTenant {
		t.Errorf("type = %q, want %q", data["type"], credential.TypeGitHubAppTenant)
	}

	if string(data["client-id"]) != "Iv1.abc123" {
		t.Errorf("client-id = %q, want %q", data["client-id"], "Iv1.abc123")
	}

	if !bytes.Equal(data["private-key"], pemBytes) {
		t.Error("private-key mismatch")
	}

	if string(data["installation-id"]) != "12345" {
		t.Errorf("installation-id = %q, want %q", data["installation-id"], "12345")
	}
}

func TestGitHubAppTenant_MarshalSecret_RoundTrip(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppTenant),
		"client-id":       []byte("Iv1.abc123"),
		"private-key":     pemBytes,
		"installation-id": []byte("12345"),
	}

	cred, err := credential.GitHubAppTenantFromSecret(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restored := cred.MarshalSecret()
	if string(restored["installation-id"]) != "12345" {
		t.Errorf("installation-id round-trip: got %q, want %q", restored["installation-id"], "12345")
	}

	if string(restored["client-id"]) != "Iv1.abc123" {
		t.Errorf("client-id round-trip: got %q, want %q", restored["client-id"], "Iv1.abc123")
	}
}

func TestGitHubAppTenant_Authenticate(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	cred := credential.NewGitHubAppTenant(app, 12345)
	httpClient := &http.Client{Transport: &tests.FakeInstallationTokenTransport{Token: "ghs_fake_token"}}

	auth, err := cred.Authenticate(t.Context(), httpClient)
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

	if basic.Password != "ghs_fake_token" {
		t.Errorf("Password = %q, want %q", basic.Password, "ghs_fake_token")
	}
}

func TestGitHubAppTenant_Authenticate_APIError(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	cred := credential.NewGitHubAppTenant(app, 12345)
	httpClient := &http.Client{Transport: &tests.FailingTransport{}}

	_, err = cred.Authenticate(t.Context(), httpClient)
	if err == nil {
		t.Fatal("expected error from failing API")
	}
}
