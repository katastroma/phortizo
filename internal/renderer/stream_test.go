//revive:disable:package-comments
package renderer

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"google.golang.org/grpc"

	"github.com/katastroma/phortizo/internal/tests"
)

func TestNewStreamFunc(t *testing.T) {
	cs := &tests.MockClientStream{}
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) { return cs, nil },
	}
	streamFn := NewStreamFunc(slog.Default(), conn)

	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	if err := streamFn(t.Context(), fs, "deploy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cs.SendMsgCount == 0 {
		t.Fatal("expected at least one message sent")
	}
}

func TestNewStreamFunc_StreamError(t *testing.T) {
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	streamFn := NewStreamFunc(slog.Default(), conn)

	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	if err := streamFn(t.Context(), fs, "deploy"); err == nil {
		t.Fatal("expected error when stream fails")
	}
}
