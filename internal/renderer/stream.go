//revive:disable:package-comments
package renderer

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"

	pb "github.com/katastroma/keleustes"
)

// chunkSize is the byte size of each RenderRequest message. This is a
// practical choice — large enough to amortize per-message overhead, small
// enough to keep memory pressure low. gRPC's default max message size is
// 4 MiB; this value is well under that limit.
const chunkSize = 32 * 1024

// stream opens a Render stream on the client and sends the source content
// from the filesystem at root as a tar archive. It closes the send side
// when done.
func stream(ctx context.Context, client pb.RendererServiceClient, fs billy.Filesystem, root string) error {
	s, err := client.Render(ctx)
	if err != nil {
		return fmt.Errorf("opening render stream: %w", err)
	}

	pr, pw := io.Pipe()

	writeErr := make(chan error, 1)
	go archiveToPipe(writeErr, pw, fs, root)

	sendErr := sendChunks(s, pr)
	pr.Close()
	archiveErr := <-writeErr

	if archiveErr != nil && !errors.Is(archiveErr, io.ErrClosedPipe) {
		return fmt.Errorf("streaming source: %w", archiveErr)
	}
	if sendErr != nil {
		return sendErr
	}

	return s.CloseSend()
}

// archiveToPipe writes a tar archive of the filesystem to the pipe and
// reports any error on the result channel.
func archiveToPipe(result chan<- error, pw *io.PipeWriter, fs billy.Filesystem, root string) {
	err := writeTar(pw, fs, root)
	result <- err

	if err != nil {
		pw.CloseWithError(err)
	} else {
		pw.Close()
	}
}

// writeTar writes the filesystem tree at root into w as a tar archive.
func writeTar(w io.Writer, fs billy.Filesystem, root string) error {
	tw := tar.NewWriter(w)
	a := &archiver{tw: tw, fs: fs, root: root}

	if err := util.Walk(fs, root, a.add); err != nil {
		return err
	}

	return tw.Close()
}

// archiver writes filesystem entries as tar archive entries with paths
// relative to root.
type archiver struct {
	tw   *tar.Writer
	fs   billy.Filesystem
	root string
}

// add is a billy walk function that writes a single file as a tar entry.
func (a *archiver) add(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	header, headerErr := tar.FileInfoHeader(info, "")
	if headerErr != nil {
		return fmt.Errorf("building tar header for %s: %w", path, headerErr)
	}
	header.Name = strings.TrimPrefix(path, a.root+"/")

	if headerErr = a.tw.WriteHeader(header); headerErr != nil {
		return fmt.Errorf("writing tar header for %s: %w", path, headerErr)
	}

	f, openErr := a.fs.Open(path)
	if openErr != nil {
		return fmt.Errorf("opening %s: %w", path, openErr)
	}
	defer f.Close()

	if _, copyErr := io.Copy(a.tw, f); copyErr != nil {
		return fmt.Errorf("writing %s to tar: %w", path, copyErr)
	}

	return nil
}

// sendChunks reads from r in chunkSize pieces and sends each as a
// RenderRequest on the stream.
func sendChunks(s pb.RendererService_RenderClient, r io.Reader) error {
	buf := make([]byte, chunkSize)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			// Copy to decouple from the reused read buffer.
			data := make([]byte, n)
			copy(data, buf[:n])

			if sendErr := s.Send(&pb.RenderRequest{Data: data}); sendErr != nil {
				return fmt.Errorf("sending chunk: %w", sendErr)
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
	}
}
