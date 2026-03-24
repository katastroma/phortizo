package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func generateTestSSHPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

func TestFromSSHKey(t *testing.T) {
	pemBytes := generateTestSSHPEM(t)

	keys, err := FromSSHKey(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys == nil {
		t.Fatal("expected non-nil keys")
	}
}

func TestFromSSHKey_InvalidPEM(t *testing.T) {
	if _, err := FromSSHKey([]byte("not a valid key")); err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}
