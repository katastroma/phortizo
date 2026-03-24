package credential

import "testing"

func TestMemoryStore_AddAndGet(t *testing.T) {
	s := NewMemoryStore()
	cred := &Credential{
		Type:  Token,
		Token: "ghp_abc123",
	}

	s.Add("ref-1", cred)

	got, err := s.Get(t.Context(), "ref-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != "ghp_abc123" {
		t.Errorf("token = %q, want %q", got.Token, "ghp_abc123")
	}
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	s := NewMemoryStore()

	if _, err := s.Get(t.Context(), "missing"); err == nil {
		t.Fatal("expected error for missing credential")
	}
}
