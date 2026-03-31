//revive:disable:package-comments
package renderer

import (
	"context"
	"log/slog"

	"github.com/go-git/go-billy/v5"
	"google.golang.org/grpc"

	pb "github.com/katastroma/keleustes"
)

// StreamFunc streams source content to the renderer service.
type StreamFunc func(ctx context.Context, fs billy.Filesystem, path string) error

// NewStreamFunc returns a StreamFunc that streams via the given connection.
func NewStreamFunc(log *slog.Logger, conn grpc.ClientConnInterface) StreamFunc {
	return func(ctx context.Context, fs billy.Filesystem, path string) error {
		return send(ctx, log, pb.NewRendererServiceClient(conn), fs, path)
	}
}
