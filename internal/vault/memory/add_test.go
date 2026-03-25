//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/auth"
	ghtoken "github.com/katastroma/phortizo/internal/auth/gh_token"
)

func TestVault_Add(t *testing.T) {
	ctx := t.Context()
	v := &Vault{credentials: make(map[string]auth.Credential)}

	secret := "ghp_abc123"
	cred := ghtoken.New(secret)

	key := "ref-1"
	v.Add(key, cred)
	if val := v.credentials[key].Token(ctx); val != secret {
		t.Fatalf("expect %v, got %v", secret, val)
	}
}
