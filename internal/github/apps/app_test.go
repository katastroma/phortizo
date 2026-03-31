package apps_test

import (
	"testing"

	"github.com/katastroma/phortizo/internal/github/apps"
	"github.com/katastroma/phortizo/internal/key"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestFromAppParameters(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestFromAppParameters_InvalidPEM(t *testing.T) {
	if _, err := apps.FromAppParameters("Iv1.abc123", []byte("not-a-pem")); err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestMarshalSecret_RoundTrip(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	app, err := apps.FromAppParameters("Iv1.abc123", pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := app.MarshalSecret()

	if string(data["client-id"]) != "Iv1.abc123" {
		t.Errorf("client-id = %q, want %q", data["client-id"], "Iv1.abc123")
	}

	if _, err = key.ParsePrivateKey(data["private-key"]); err != nil {
		t.Fatalf("marshaled private-key is not valid PEM: %v", err)
	}
}
