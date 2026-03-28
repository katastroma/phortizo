//revive:disable:package-comments
package key

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// ParsePrivateKey parses a PKCS1 private key from PEM or base64-encoded PEM.
func ParsePrivateKey(raw []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(string(raw))
		if err != nil {
			return nil, fmt.Errorf("input is neither PEM nor valid base64")
		}

		block, _ = pem.Decode(decoded)
		if block == nil {
			return nil, fmt.Errorf("no PEM block found after base64 decoding")
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing PKCS1 private key: %w", err)
	}

	return key, nil
}

// MarshalPrivateKey encodes an RSA private key as PEM-encoded PKCS1 bytes.
func MarshalPrivateKey(k *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k),
	})
}
