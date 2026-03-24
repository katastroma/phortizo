package key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func generateTestPEM(t *testing.T) []byte {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}

func TestParsePrivateKey_Valid(t *testing.T) {
	pemBytes := generateTestPEM(t)

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
