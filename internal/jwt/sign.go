//revive:disable:package-comments
package jwt

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sign creates a signed JWT from an issuer and a private key
func Sign(
	issuer string,
	issuedAt time.Time,
	expiresAt time.Time,
	key *rsa.PrivateKey,
) (string, error) {
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}
