//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/auth"
	ghtoken "github.com/katastroma/phortizo/internal/auth/gh_token"
)

func TestVault_Withdraw(t *testing.T) {
	ctx := t.Context()
	secret := "ghp_abc123"
	token := ghtoken.New(secret)
	v := &Vault{credentials: map[string]auth.Credential{"ref-1": token}}

	cred, err := v.Withdraw(ctx, "ref-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if val := cred.Token(ctx); val != secret {
		t.Fatalf("expect %v, got %v", secret, val)

	}
}

func TestVault_Withdraw_NotFound(t *testing.T) {
	v := &Vault{credentials: make(map[string]auth.Credential)}

	if _, err := v.Withdraw(t.Context(), "missing"); err == nil {
		t.Fatal("expected error for missing credential")
	}
}
