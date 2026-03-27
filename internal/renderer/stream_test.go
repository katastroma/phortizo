package renderer

import (
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

func createTestFile(t *testing.T, fs billy.Filesystem, path, content string) {
	t.Helper()
	f, err := fs.Create(path)
	if err != nil {
		t.Fatalf("creating %s: %v", path, err)
	}
	f.Write([]byte(content))
	f.Close()
}

func TestStream(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.sent) == 0 {
		t.Fatal("expected at least one message sent")
	}
}

func TestStream_MultipleFiles(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/a.yaml", "a")
	createTestFile(t, fs, "deploy/b.yaml", "b")

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.sent) < 2 {
		t.Errorf("expected at least 2 messages, got %d", len(stream.sent))
	}
}

func TestStream_RenderOpenError(t *testing.T) {
	fs := memfs.New()
	client := &mockClient{err: fmt.Errorf("connection refused")}

	if err := Stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when render stream fails to open")
	}
}

func TestStream_SendError(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	stream := &mockStream{sendErr: fmt.Errorf("send failed")}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when send fails")
	}
}

func TestStream_CloseSendError(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	stream := &mockStream{closeSendErr: fmt.Errorf("close failed")}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when close send fails")
	}
}

func TestStream_FileOpenError(t *testing.T) {
	backing := memfs.New()
	createTestFile(t, backing, "deploy/values.yaml", "key: value")
	fs := &openErrorFS{Filesystem: backing}

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when file open fails")
	}
}

func TestStream_FileReadError(t *testing.T) {
	backing := memfs.New()
	createTestFile(t, backing, "deploy/values.yaml", "key: value")
	fs := &errorFS{Filesystem: backing}

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when file read fails")
	}
}

func TestStream_WalkError(t *testing.T) {
	fs := memfs.New()

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestStream_EmptyDirectory(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("empty", 0o755)

	stream := &mockStream{}
	client := &mockClient{stream: stream}

	if err := Stream(t.Context(), client, fs, "empty"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.sent) != 0 {
		t.Errorf("expected 0 messages for empty dir, got %d", len(stream.sent))
	}
}
