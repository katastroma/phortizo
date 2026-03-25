package key

import (
	"encoding/pem"
	"testing"

	"github.com/katastroma/phortizo/internal/tests"
)

func TestParsePrivateKey_Valid(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	key, err := ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestParsePrivateKey_NoPEMBlock(t *testing.T) {
	if _, err := ParsePrivateKey([]byte("not a pem block")); err == nil {
		t.Fatal("expected error for non-PEM input")
	}
}

func TestParsePrivateKey_InvalidDER(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid der data"),
	})

	if _, err := ParsePrivateKey(pemBytes); err == nil {
		t.Fatal("expected error for invalid DER data")
	}
}
