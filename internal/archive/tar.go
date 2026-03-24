//revive:disable:package-comments
package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
)

// FromBillyFS creates a tar+gzip archive of the subtree at root within fs.
// Paths in the archive are relative to root.
func FromBillyFS(fs billy.Filesystem, root string) (io.Reader, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	w := &tarWalker{fs: fs, tw: tw, root: root}
	if err := util.Walk(fs, root, w.addFile); err != nil {
		return nil, fmt.Errorf("walking %s: %w", root, err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing tar writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("closing gzip writer: %w", err)
	}

	return &buf, nil
}

type tarWalker struct {
	fs   billy.Filesystem
	tw   *tar.Writer
	root string
}

func (w *tarWalker) addFile(filePath string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	relPath, err := relativePath(w.root, filePath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("creating tar header for %s: %w", relPath, err)
	}
	header.Name = relPath

	if err = w.tw.WriteHeader(header); err != nil {
		return fmt.Errorf("writing tar header for %s: %w", relPath, err)
	}

	f, err := w.fs.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(w.tw, f); err != nil {
		return fmt.Errorf("writing %s to tar: %w", relPath, err)
	}

	return nil
}

func relativePath(base, target string) (string, error) {
	base = strings.TrimSuffix(base, "/")
	if base == "" || base == "." || base == "/" {
		return target, nil
	}
	prefix := base + "/"
	if !strings.HasPrefix(target, prefix) {
		return "", fmt.Errorf("%s is not under %s", target, base)
	}
	return strings.TrimPrefix(target, prefix), nil
}
