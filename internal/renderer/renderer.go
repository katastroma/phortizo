//revive:disable:package-comments
package renderer

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/katastroma/keleustes"
)

// Type identifies which renderer backend should process the source
type Type string

const (
	// Helm indicates a Helm chart (Chart.yaml present)
	Helm Type = "helm"
	// Kustomize indicates a kustomization (kustomization.yaml present)
	Kustomize Type = "kustomize"
	// Raw indicates plain YAML manifests with no renderer framework
	Raw Type = "raw"
)

// Streamer streams source content to a renderer service
type Streamer interface {
	Stream(ctx context.Context, fs billy.Filesystem, path, addr string) error
}

// Client connects to a renderer service and streams source content
type Client struct{}

// Stream dials the renderer at addr, streams the filesystem content, and
// closes the connection.
func (Client) Stream(ctx context.Context, fs billy.Filesystem, path, addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connecting to renderer at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewRendererServiceClient(conn)
	return stream(ctx, client, fs, path)
}
