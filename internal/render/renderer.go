//revive:disable:package-comments
package render

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/katastroma/keleustes"
)

// Renderer connects to a renderer service and streams source content.
type Renderer struct{}

// Render dials the renderer at addr, streams the filesystem content, and
// closes the connection.
func (Renderer) Render(ctx context.Context, fs billy.Filesystem, path, addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connecting to renderer at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewRendererServiceClient(conn)
	return Stream(ctx, client, fs, path)
}
