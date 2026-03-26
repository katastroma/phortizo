//revive:disable:package-comments
package retriever

import (
	"context"
	"log/slog"
	"testing"
	"time"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/tracequery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubQuerier struct {
	attrs tracequery.Attributes
	err   error
}

func (q *stubQuerier) SpanAttributes(_ context.Context, _, _ string) (tracequery.Attributes, error) {
	return q.attrs, q.err
}

func TestRetrieve_Unimplemented(t *testing.T) {
	handler := New(slog.Default(), nil, nil, nil, 10*time.Minute, 3)

	_, err := handler.Retrieve(t.Context(), &pb.RetrieveRequest{})
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}
