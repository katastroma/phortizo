package auth

import "testing"

func TestFromBasicAuth(t *testing.T) {
	got := FromBasicAuth("user", "pass")

	if got.Username != "user" {
		t.Errorf("username = %q, want %q", got.Username, "user")
	}
	if got.Password != "pass" {
		t.Errorf("password = %q, want %q", got.Password, "pass")
	}
}
