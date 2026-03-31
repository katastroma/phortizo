package key_test

import (
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/katastroma/phortizo/internal/key"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestParsePrivateKey_Valid(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	key, err := key.ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestParsePrivateKey_Base64EncodedPEM(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)
	encoded := []byte(base64.StdEncoding.EncodeToString(pemBytes))

	key, err := key.ParsePrivateKey(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestParsePrivateKey_NoPEMBlock(t *testing.T) {
	if _, err := key.ParsePrivateKey([]byte("not a pem block")); err == nil {
		t.Fatal("expected error for non-PEM input")
	}
}

func TestParsePrivateKey_ValidBase64ButNotPEM(t *testing.T) {
	encoded := []byte(base64.StdEncoding.EncodeToString([]byte("just some text")))

	if _, err := key.ParsePrivateKey(encoded); err == nil {
		t.Fatal("expected error for base64 that does not contain PEM")
	}
}

func TestParsePrivateKey_Base64EncodedPEM_InvalidDER(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid der data"),
	})
	encoded := []byte(base64.StdEncoding.EncodeToString(pemBytes))

	if _, err := key.ParsePrivateKey(encoded); err == nil {
		t.Fatal("expected error for invalid DER data in base64-encoded PEM")
	}
}

func TestParsePrivateKey_InvalidDER(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid der data"),
	})

	if _, err := key.ParsePrivateKey(pemBytes); err == nil {
		t.Fatal("expected error for invalid DER data")
	}
}

func TestMarshalPrivateKey_RoundTrip(t *testing.T) {
	pemBytes := tests.GenerateRSAPEM(t)

	original, err := key.ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	marshaled := key.MarshalPrivateKey(original)

	restored, err := key.ParsePrivateKey(marshaled)
	if err != nil {
		t.Fatalf("unexpected error parsing marshaled key: %v", err)
	}

	if !original.Equal(restored) {
		t.Fatal("round-tripped key does not match original")
	}
}
