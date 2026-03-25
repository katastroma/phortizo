package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func sign(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func validPayload() []byte {
	return []byte(`{
		"ref": "refs/heads/main",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [{"added": ["deploy/values.yaml"], "removed": [], "modified": []}]
	}`)
}
