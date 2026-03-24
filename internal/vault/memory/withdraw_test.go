//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/vault"
)

func TestVault_Withdraw(t *testing.T) {
	v := &Vault{credentials: map[string]*vault.Credential{
		"ref-1": {Type: vault.Token, Token: "ghp_abc123"},
	}}

	got, err := v.Withdraw(t.Context(), "ref-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != "ghp_abc123" {
		t.Errorf("token = %q, want %q", got.Token, "ghp_abc123")
	}
}

func TestVault_Withdraw_NotFound(t *testing.T) {
	v := &Vault{credentials: make(map[string]*vault.Credential)}

	if _, err := v.Withdraw(t.Context(), "missing"); err == nil {
		t.Fatal("expected error for missing credential")
	}
}
