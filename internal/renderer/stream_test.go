package renderer

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/katastroma/phortizo/internal/tests"
)

func createTestFile(t *testing.T, fs billy.Filesystem, path, content string) {
	t.Helper()
	f, err := fs.Create(path)
	if err != nil {
		t.Fatalf("creating %s: %v", path, err)
	}
	if _, err = f.Write([]byte(content)); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	if err = f.Close(); err != nil {
		t.Fatalf("closing %s: %v", path, err)
	}
}

func extractTar(t *testing.T, ms *tests.MockRenderStream) map[string]string {
	t.Helper()

	var buf bytes.Buffer
	for _, req := range ms.Sent {
		buf.Write(req.GetData())
	}

	files := make(map[string]string)
	tr := tar.NewReader(&buf)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar: %v", err)
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("reading tar entry %s: %v", header.Name, err)
		}
		files[header.Name] = string(content)
	}

	return files
}

func TestStream(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := extractTar(t, ms)
	content, ok := files["values.yaml"]
	if !ok {
		t.Fatal("expected values.yaml in tar archive")
	}
	if content != "key: value" {
		t.Errorf("expected %q, got %q", "key: value", content)
	}
}

func TestStream_MultipleFiles(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/a.yaml", "a")
	createTestFile(t, fs, "deploy/b.yaml", "b")

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := extractTar(t, ms)
	if len(files) != 2 {
		t.Errorf("expected 2 files in tar, got %d", len(files))
	}
	if files["a.yaml"] != "a" {
		t.Errorf("expected %q for a.yaml, got %q", "a", files["a.yaml"])
	}
	if files["b.yaml"] != "b" {
		t.Errorf("expected %q for b.yaml, got %q", "b", files["b.yaml"])
	}
}

func TestStream_RenderOpenError(t *testing.T) {
	fs := memfs.New()
	client := &tests.MockRendererClient{Err: fmt.Errorf("connection refused")}

	if err := stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when render stream fails to open")
	}
}

func TestStream_SendError(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	ms := &tests.MockRenderStream{SendErr: fmt.Errorf("send failed")}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when send fails")
	}
}

func TestStream_CloseSendError(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "deploy/values.yaml", "key: value")

	ms := &tests.MockRenderStream{CloseSendErr: fmt.Errorf("close failed")}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when close send fails")
	}
}

func TestStream_FileOpenError(t *testing.T) {
	backing := memfs.New()
	createTestFile(t, backing, "deploy/values.yaml", "key: value")
	fs := &tests.OpenErrorFS{Filesystem: backing}

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when file open fails")
	}
}

func TestStream_FileReadError(t *testing.T) {
	backing := memfs.New()
	createTestFile(t, backing, "deploy/values.yaml", "key: value")
	fs := &tests.ErrorFS{Filesystem: backing}

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "deploy"); err == nil {
		t.Fatal("expected error when file read fails")
	}
}

func TestStream_WalkError(t *testing.T) {
	fs := memfs.New()

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestStream_EmptyDirectory(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("empty", 0o755)

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "empty"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := extractTar(t, ms)
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty dir, got %d", len(files))
	}
}

func TestStream_RootPath(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "values.yaml", "key: value")

	ms := &tests.MockRenderStream{}
	client := &tests.MockRendererClient{Stream: ms}

	if err := stream(t.Context(), client, fs, "."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := extractTar(t, ms)
	content, ok := files["values.yaml"]
	if !ok {
		t.Fatal("expected values.yaml in tar archive")
	}
	if content != "key: value" {
		t.Errorf("expected %q, got %q", "key: value", content)
	}
}

func TestWriteTar(t *testing.T) {
	fs := memfs.New()
	createTestFile(t, fs, "root/a.yaml", "hello")

	var buf bytes.Buffer
	if err := writeTar(&buf, fs, "root"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := make(map[string]string)
	tr := tar.NewReader(&buf)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar: %v", err)
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("reading entry: %v", err)
		}
		files[header.Name] = string(content)
	}

	if files["a.yaml"] != "hello" {
		t.Errorf("expected %q, got %q", "hello", files["a.yaml"])
	}
}

func TestWriteTar_CloseError(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("empty", 0o755)

	if err := writeTar(&tests.ErrWriter{}, fs, "empty"); err == nil {
		t.Fatal("expected error when tar close fails")
	}
}

func TestArchiver_Add_WalkError(t *testing.T) {
	tw := tar.NewWriter(&bytes.Buffer{})
	a := &archiver{tw: tw, fs: memfs.New(), root: "."}

	walkErr := fmt.Errorf("walk error")
	if err := a.add("path", nil, walkErr); err != walkErr {
		t.Fatalf("expected walk error, got %v", err)
	}
}

func TestArchiver_Add_HeaderError(t *testing.T) {
	tw := tar.NewWriter(&bytes.Buffer{})
	a := &archiver{tw: tw, fs: memfs.New(), root: "."}

	if err := a.add("irregular", tests.IrregularFileInfo{}, nil); err == nil {
		t.Fatal("expected error for irregular file type")
	}
}

func TestArchiver_Add_WriteHeaderError(t *testing.T) {
	tw := tar.NewWriter(&tests.ErrWriter{})
	fs := memfs.New()
	createTestFile(t, fs, "test.yaml", "content")

	info, err := fs.Stat("test.yaml")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	a := &archiver{tw: tw, fs: fs, root: "."}
	if err := a.add("test.yaml", info, nil); err == nil {
		t.Fatal("expected error when write header fails")
	}
}
