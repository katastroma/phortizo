package registration

import "testing"

func TestMemoryStore_AddAndGetByID(t *testing.T) {
	s := NewMemoryStore()
	reg := &Registration{
		ID:       "abc",
		TenantID: "acme",
		Secret:   []byte("secret"),
	}

	s.Add(reg)

	got, err := s.GetByID(t.Context(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TenantID != "acme" {
		t.Errorf("tenant_id = %q, want %q", got.TenantID, "acme")
	}
}

func TestMemoryStore_GetByID_NotFound(t *testing.T) {
	s := NewMemoryStore()

	if _, err := s.GetByID(t.Context(), "missing"); err == nil {
		t.Fatal("expected error for missing registration")
	}
}
