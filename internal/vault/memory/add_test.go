//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/vault"
)

func TestVault_Add(t *testing.T) {
	v := &Vault{credentials: make(map[string]*vault.Credential)}

	cred := &vault.Credential{
		Type:  vault.Token,
		Token: "ghp_abc123",
	}

	v.Add("ref-1", cred)

	got, ok := v.credentials["ref-1"]
	if !ok {
		t.Fatal("credential not found in map after Add")
	}
	if got.Token != "ghp_abc123" {
		t.Errorf("token = %q, want %q", got.Token, "ghp_abc123")
	}
}
