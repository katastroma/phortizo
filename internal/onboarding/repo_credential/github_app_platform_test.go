package credential_test

import (
	"net/http"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/katastroma/phortizo/internal/github"
	"github.com/katastroma/phortizo/internal/github/apps"
	credential "github.com/katastroma/phortizo/internal/onboarding/repo_credential"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestGitHubAppPlatform_FromSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppPlatform),
		"installation-id": []byte("67890"),
	}

	cred, err := credential.GitHubAppPlatformFromSecret(data, platformApp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
}

func TestGitHubAppPlatform_FromSecret_MissingInstallationID(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	data := map[string][]byte{
		"type": []byte(credential.TypeGitHubAppPlatform),
	}

	_, err = credential.GitHubAppPlatformFromSecret(data, platformApp)
	if err == nil {
		t.Fatal("expected error for missing installation-id")
	}
}

func TestGitHubAppPlatform_FromSecret_InvalidInstallationID(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	data := map[string][]byte{
		"type":            []byte(credential.TypeGitHubAppPlatform),
		"installation-id": []byte("not-a-number"),
	}

	_, err = credential.GitHubAppPlatformFromSecret(data, platformApp)
	if err == nil {
		t.Fatal("expected error for invalid installation-id")
	}
}

func TestGitHubAppPlatform_MarshalSecret(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	cred := credential.NewGitHubAppPlatform(platformApp, 67890)
	data := cred.MarshalSecret()

	if string(data["type"]) != credential.TypeGitHubAppPlatform {
		t.Errorf("type = %q, want %q", data["type"], credential.TypeGitHubAppPlatform)
	}

	if string(data["installation-id"]) != "67890" {
		t.Errorf("installation-id = %q, want %q", data["installation-id"], "67890")
	}
}

func TestGitHubAppPlatform_Authenticate(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	cred := credential.NewGitHubAppPlatform(platformApp, 67890)
	httpClient := &http.Client{Transport: &tests.FakeInstallationTokenTransport{Token: "ghs_platform_token"}}

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

	if basic.Password != "ghs_platform_token" {
		t.Errorf("Password = %q, want %q", basic.Password, "ghs_platform_token")
	}
}

func TestGitHubAppPlatform_Authenticate_APIError(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	platformApp, err := apps.FromAppParameters("Iv1.platform", pemBytes)
	if err != nil {
		t.Fatalf("creating platform app: %v", err)
	}

	cred := credential.NewGitHubAppPlatform(platformApp, 67890)
	httpClient := &http.Client{Transport: &tests.FailingTransport{}}

	_, err = cred.Authenticate(t.Context(), httpClient)
	if err == nil {
		t.Fatal("expected error from failing API")
	}
}
