//revive:disable:package-comments
package memory

import "testing"

func TestNew(t *testing.T) {
	r := New()

	if r.registrations == nil {
		t.Fatal("registrations map not initialized")
	}
}
