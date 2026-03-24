package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestSignature_Valid(t *testing.T) {
	payload := []byte(`{"ref":"refs/heads/main"}`)
	secret := []byte("test-secret")
	header := sign(payload, secret)

	if err := Signature(payload, secret, header); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestSignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"ref":"refs/heads/main"}`)
	header := sign(payload, []byte("correct-secret"))

	if err := Signature(payload, []byte("wrong-secret"), header); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestSignature_TamperedPayload(t *testing.T) {
	secret := []byte("test-secret")
	header := sign([]byte(`{"ref":"refs/heads/main"}`), secret)

	if err := Signature([]byte(`{"ref":"refs/heads/evil"}`), secret, header); err == nil {
		t.Fatal("expected error for tampered payload")
	}
}

func TestSignature_MissingPrefix(t *testing.T) {
	if err := Signature([]byte("payload"), []byte("secret"), "abc123"); err == nil {
		t.Fatal("expected error for missing sha256= prefix")
	}
}

func TestSignature_InvalidHex(t *testing.T) {
	if err := Signature([]byte("payload"), []byte("secret"), "sha256=notvalidhex!"); err == nil {
		t.Fatal("expected error for invalid hex")
	}
}
