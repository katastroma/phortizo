//revive:disable:package-comments
package grpc

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/katastroma/keleustes"

	"github.com/katastroma/phortizo/internal/render"
)

// StreamToRenderer streams to the renderer service
func StreamToRenderer(
	ctx context.Context,
	tracer trace.Tracer,
	fs billy.Filesystem,
	path, address string,
) error {
	ctx, span := tracer.Start(ctx, "pipeline.stream")
	defer span.End()

	connOpts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(address, connOpts)
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
