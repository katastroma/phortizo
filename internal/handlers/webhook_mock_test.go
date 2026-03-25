package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/katastroma/phortizo/internal/registration"
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

func testRegistration() *registration.Record {
	return &registration.Record{
		ID:            "reg-1",
		TenantID:      "acme",
		Secret:        []byte("test-secret"),
		CredentialRef: "cred-1",
		WatchTargets: []registration.WatchTarget{
			{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		},
	}
}
