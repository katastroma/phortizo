//revive:disable:package-comments
package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/katastroma/phortizo/internal/registration"
)

// Signature checks that the HMAC-SHA256 signature in the X-Hub-Signature-256
// header matches the payload signed with secret.
func Signature(payload, secret registration.WebhookSecret, signatureHeader string) error {
	sig, found := strings.CutPrefix(signatureHeader, "sha256=")
	if !found {
		return fmt.Errorf("missing sha256= prefix in signature header")
	}

	decoded, err := hex.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("decoding signature hex: %w", err)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(decoded, expected) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
