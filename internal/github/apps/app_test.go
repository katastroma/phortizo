package apps

import (
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/katastroma/phortizo/internal/key"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestFromAppParameters(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	app, err := FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestFromAppParameters_InvalidPEM(t *testing.T) {
	_, err := FromAppParameters("Iv1.abc123", []byte("not-a-pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestMarshalSecret_RoundTrip(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := app.MarshalSecret()

	if string(data["client-id"]) != "Iv1.abc123" {
		t.Errorf("client-id = %q, want %q", data["client-id"], "Iv1.abc123")
	}

	_, err = key.ParsePrivateKey(data["private-key"])
	if err != nil {
		t.Fatalf("marshaled private-key is not valid PEM: %v", err)
	}
}

func TestExchangeInstallationToken(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	pk, err := key.ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("parsing key: %v", err)
	}

	app := &App{clientID: "Iv1.abc123", privateKey: pk}
	httpClient := &http.Client{Transport: &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"}}

	token, err := app.ExchangeInstallationToken(t.Context(), httpClient, 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "ghs_test_token" {
		t.Errorf("token = %q, want %q", token, "ghs_test_token")
	}
}

func TestExchangeInstallationToken_APIError(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	pk, err := key.ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("parsing key: %v", err)
	}

	app := &App{clientID: "Iv1.abc123", privateKey: pk}
	httpClient := &http.Client{Transport: &tests.FailingTransport{}}

	_, err = app.ExchangeInstallationToken(t.Context(), httpClient, 12345)
	if err == nil {
		t.Fatal("expected error from failing API")
	}
}

func TestExchangeInstallationToken_NilKey(t *testing.T) {
	app := &App{clientID: "Iv1.abc123", privateKey: nil}
	httpClient := &http.Client{Transport: &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"}}

	_, err := app.ExchangeInstallationToken(t.Context(), httpClient, 12345)
	if err == nil {
		t.Fatal("expected error from nil private key")
	}
}

func TestExchangeInstallationToken_SignError(t *testing.T) {
	app := &App{clientID: "Iv1.abc123", privateKey: &rsa.PrivateKey{}}
	httpClient := &http.Client{Transport: &tests.FakeInstallationTokenTransport{Token: "ghs_test_token"}}

	_, err := app.ExchangeInstallationToken(t.Context(), httpClient, 12345)
	if err == nil {
		t.Fatal("expected error from invalid private key")
	}
}
