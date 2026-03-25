//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	pb "github.com/katastroma/keleustes"
	"github.com/katastroma/phortizo/internal/render"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (r *Runner) stream(ctx context.Context, fs billy.Filesystem, path, address string) error {
	ctx, span := tracer.Start(ctx, "pipeline.stream")
	defer span.End()

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("connecting to renderer at %s: %w", address, err)
	}
	defer conn.Close()

	client := pb.NewRendererServiceClient(conn)
	if err := render.Stream(ctx, client, fs, path); err != nil {
		span.RecordError(err)
		return fmt.Errorf("streaming to renderer: %w", err)
	}
	return nil
}
