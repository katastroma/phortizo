package handler

import (
	"log/slog"
	"testing"

	pb "github.com/katastroma/naukleros"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetrieve_Unimplemented(t *testing.T) {
	handler := NewRetriever(slog.Default(), nil, nil, nil)

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
