package tempo

import "testing"

func TestNew(t *testing.T) {
	q := New(&mockQuerierClient{})
	if q.client == nil {
		t.Fatal("expected non-nil client")
	}
}
