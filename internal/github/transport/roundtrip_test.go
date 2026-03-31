package transport

import (
	"net/http"
	"testing"
	"time"

	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestRoundTrip_CachesToken(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	counter := &countingTransport{inner: &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"}}
	auth := NewInstallationTokenAuth(app, 12345, counter)

	req1, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/meta", nil)
	req2, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/meta", nil)

	resp1, err := auth.RoundTrip(req1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	resp1.Body.Close()

	resp2, err := auth.RoundTrip(req2)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	resp2.Body.Close()

	// First request: exchange (2 round trips: token exchange + actual request)
	// Second request: cached (1 round trip: actual request only)
	if counter.count != 3 {
		t.Errorf("round trips = %d, want %d", counter.count, 3)
	}
}

func TestRoundTrip_RefreshesExpiredToken(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	auth := NewInstallationTokenAuth(app, 12345, &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"})

	auth.mu.Lock()
	auth.token = "stale"
	auth.expiresAt = time.Now().Add(-1 * time.Minute)
	auth.mu.Unlock()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/meta", nil)
	resp, err := auth.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	auth.mu.Lock()
	token := auth.token
	auth.mu.Unlock()

	if token != "ghs_test_token" {
		t.Errorf("token = %q, want %q", token, "ghs_test_token")
	}
}

func TestRoundTrip_ExchangeError(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	auth := NewInstallationTokenAuth(app, 12345, &tests.FailingTransport{})

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.github.com/meta", nil)
	_, err = auth.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error from failing exchange")
	}
}

func TestNewInstallationTokenAuth_NilBase(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}

	auth := NewInstallationTokenAuth(app, 12345, nil)
	if auth.base != http.DefaultTransport {
		t.Error("expected http.DefaultTransport as base")
	}
}

type countingTransport struct {
	inner http.RoundTripper
	count int
}

func (c *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.count++
	return c.inner.RoundTrip(req)
}
