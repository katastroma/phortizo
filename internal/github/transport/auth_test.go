package transport_test

import (
	"net/http"
	"testing"

	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/github/transport"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestRoundTrip(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	auth := transport.NewInstallationTokenAuth(app, 12345, &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"})

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/meta", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := auth.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}
