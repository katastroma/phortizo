package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

func TestSign_ProducesValidJWT(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}

	now := time.Now()
	issuedAt := now.Add(-60 * time.Second)
	expiresAt := now.Add(10 * time.Minute)

	tokenString, err := Sign("test-app-id", issuedAt, expiresAt, priv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := gojwt.ParseWithClaims(tokenString, &gojwt.RegisteredClaims{}, func(t *gojwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("parsing token: %v", err)
	}

	claims := token.Claims.(*gojwt.RegisteredClaims)
	if claims.Issuer != "test-app-id" {
		t.Errorf("issuer = %q, want %q", claims.Issuer, "test-app-id")
	}
}

func TestSign_NilKeyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil key")
		}
	}()

	Sign("app-id", time.Now(), time.Now().Add(time.Minute), nil)
}
