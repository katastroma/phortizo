//revive:disable:package-comments
package retriever_test

import (
	"log/slog"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/katastroma/naukleros"

	"github.com/katastroma/phortizo/internal/retriever"
)

func TestNew(t *testing.T) {
	handler := retriever.New(
		slog.Default(), nil, fake.NewSimpleClientset(),
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestRetrieve_MissingSourceTarget(t *testing.T) {
	handler := retriever.New(
		slog.Default(), nil, fake.NewSimpleClientset(),
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Retrieve(t.Context(), &pb.RetrieveRequest{
		Namespace:      "tenant-a",
		SourceTargetId: "nonexistent",
	}); err == nil {
		t.Fatal("expected error for missing source target")
	}
}
