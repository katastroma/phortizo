//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/registration"
)

func TestRegistrar_GetByID(t *testing.T) {
	r := &Registrar{registrations: map[string]*registration.Record{
		"abc": {ID: "abc", TenantID: "acme", Secret: []byte("secret")},
	}}

	got, err := r.GetByID(t.Context(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TenantID != "acme" {
		t.Errorf("tenant_id = %q, want %q", got.TenantID, "acme")
	}
}

func TestRegistrar_GetByID_NotFound(t *testing.T) {
	r := &Registrar{registrations: make(map[string]*registration.Record)}

	if _, err := r.GetByID(t.Context(), "missing"); err == nil {
		t.Fatal("expected error for missing registration")
	}
}
