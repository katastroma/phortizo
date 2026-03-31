package tempo_test

import (
	"testing"

	"github.com/katastroma/phortizo/internal/tracing/tempo"
)

func TestNew(t *testing.T) {
	tracer := tempo.New(nil)
	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}

	if tracer.GetClient() != nil {
		t.Fatal("expected nil client for nil input")
	}
}
