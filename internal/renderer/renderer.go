//revive:disable:package-comments
package renderer

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"google.golang.org/grpc"

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

// StreamFunc returns a closure that streams source content to a renderer
// service via pre-established gRPC connections keyed by service address.
func StreamFunc(
	connections map[string]grpc.ClientConnInterface,
) func(ctx context.Context, fs billy.Filesystem, path, addr string) error {
	return func(ctx context.Context, fs billy.Filesystem, path, addr string) error {
		conn, ok := connections[addr]
		if !ok {
			return fmt.Errorf("no connection for renderer at %s", addr)
		}

		return stream(ctx, pb.NewRendererServiceClient(conn), fs, path)
	}
}
