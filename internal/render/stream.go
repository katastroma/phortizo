//revive:disable:package-comments
package render

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"

	pb "github.com/katastroma/keleustes"
)

// Stream opens a Render stream on the client and sends the source content
// from the filesystem at root. It closes the send side when done.
func Stream(ctx context.Context, client pb.RendererServiceClient, fs billy.Filesystem, root string) error {
	stream, err := client.Render(ctx)
	if err != nil {
		return fmt.Errorf("opening render stream: %w", err)
	}

	w := &walker{fs: fs, stream: stream}
	if err := util.Walk(fs, root, w.send); err != nil {
		return fmt.Errorf("streaming source: %w", err)
	}

	return stream.CloseSend()
}

type walker struct {
	fs     billy.Filesystem
	stream pb.RendererService_RenderClient
}

func (w *walker) send(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	f, err := w.fs.Open(path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if sendErr := w.stream.Send(&pb.RenderRequest{Data: buf[:n]}); sendErr != nil {
				return fmt.Errorf("sending %s: %w", path, sendErr)
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
	}
}
