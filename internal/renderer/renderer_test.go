//revive:disable:package-comments
package renderer

import (
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"google.golang.org/grpc"

	"github.com/katastroma/phortizo/internal/tests"
)

func TestStreamFunc(t *testing.T) {
	cs := &tests.MockClientStream{}
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) { return cs, nil },
	}
	connections := map[string]grpc.ClientConnInterface{"helm:8080": conn}
	streamFn := StreamFunc(connections)

	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	if err := streamFn(t.Context(), fs, "deploy", "helm:8080"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cs.SendMsgCount == 0 {
		t.Fatal("expected at least one message sent")
	}
}

func TestStreamFunc_NoConnection(t *testing.T) {
	connections := map[string]grpc.ClientConnInterface{}
	streamFn := StreamFunc(connections)

	fs := memfs.New()

	if err := streamFn(t.Context(), fs, "deploy", "unknown:8080"); err == nil {
		t.Fatal("expected error for missing connection")
	}
}

func TestStreamFunc_StreamError(t *testing.T) {
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	connections := map[string]grpc.ClientConnInterface{"helm:8080": conn}
	streamFn := StreamFunc(connections)

	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	if err := streamFn(t.Context(), fs, "deploy", "helm:8080"); err == nil {
		t.Fatal("expected error when stream fails")
	}
}
