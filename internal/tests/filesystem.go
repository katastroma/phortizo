//revive:disable:package-comments
package tests

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-billy/v5"
)

// ErrorFile implements billy.File with a Read that always errors.
type ErrorFile struct{ billy.File }

// Read always returns a disk error.
func (f *ErrorFile) Read(_ []byte) (int, error) { return 0, fmt.Errorf("disk error") }

// Close is a no-op.
func (f *ErrorFile) Close() error { return nil }

// ErrorFS wraps a real filesystem but returns an ErrorFile on Open.
type ErrorFS struct{ billy.Filesystem }

// Open returns an ErrorFile regardless of path.
func (fs *ErrorFS) Open(_ string) (billy.File, error) {
	return &ErrorFile{}, nil
}

// Stat delegates to the wrapped filesystem.
func (fs *ErrorFS) Stat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Stat(name)
}

// OpenErrorFS wraps a real filesystem but errors on Open.
type OpenErrorFS struct{ billy.Filesystem }

// Open always returns an error.
func (fs *OpenErrorFS) Open(_ string) (billy.File, error) {
	return nil, fmt.Errorf("open denied")
}

// Stat delegates to the wrapped filesystem.
func (fs *OpenErrorFS) Stat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Stat(name)
}

// ErrWriter is an io.Writer that always returns an error.
type ErrWriter struct{}

// Write always returns a write error.
func (ErrWriter) Write(_ []byte) (int, error) { return 0, fmt.Errorf("write error") }

// IrregularFileInfo implements os.FileInfo with ModeIrregular, which
// tar.FileInfoHeader cannot represent.
type IrregularFileInfo struct{}

// Name returns a placeholder name.
func (IrregularFileInfo) Name() string { return "irregular" }

// Size returns 0.
func (IrregularFileInfo) Size() int64 { return 0 }

// Mode returns ModeIrregular, which causes tar.FileInfoHeader to error.
func (IrregularFileInfo) Mode() os.FileMode { return os.ModeIrregular }

// ModTime returns the zero time.
func (IrregularFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir returns false.
func (IrregularFileInfo) IsDir() bool { return false }

// Sys returns nil.
func (IrregularFileInfo) Sys() any { return nil }
