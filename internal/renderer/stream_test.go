//revive:disable:package-comments
package renderer_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"google.golang.org/grpc"

	"github.com/katastroma/keleustes"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestNewStreamFunc(t *testing.T) {
	cs := &tests.MockClientStream{Ctx: t.Context()}
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) { return cs, nil },
	}
	streamFn := renderer.NewStreamFunc(slog.Default(), conn, 32*1024)

	if err := streamFn(t.Context(), memfs.New(), ".", keleustes.RendererType_RENDERER_TYPE_PLAIN); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStreamFunc_StreamError(t *testing.T) {
	conn := &tests.MockClientConn{
		NewStreamFn: func() (grpc.ClientStream, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	streamFn := renderer.NewStreamFunc(slog.Default(), conn, 32*1024)

	if err := streamFn(t.Context(), memfs.New(), ".", keleustes.RendererType_RENDERER_TYPE_PLAIN); err == nil {
		t.Fatal("expected error when stream fails")
	}
}
