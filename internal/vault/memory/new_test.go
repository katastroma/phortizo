//revive:disable:package-comments
package memory

import "testing"

func TestNew(t *testing.T) {
	v := New()

	if v.credentials == nil {
		t.Fatal("credentials map not initialized")
	}
}
