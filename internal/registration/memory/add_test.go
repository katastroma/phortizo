//revive:disable:package-comments
package memory

import (
	"testing"

	"github.com/katastroma/phortizo/internal/registration"
)

func TestRegistrar_Add(t *testing.T) {
	r := &Registrar{registrations: make(map[string]*registration.Record)}

	reg := &registration.Record{
		ID:       "abc",
		TenantID: "acme",
		Secret:   []byte("secret"),
	}

	r.Add(reg)

	got, ok := r.registrations["abc"]
	if !ok {
		t.Fatal("record not found in map after Add")
	}
	if got.TenantID != "acme" {
		t.Errorf("tenant_id = %q, want %q", got.TenantID, "acme")
	}
}
