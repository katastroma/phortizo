//revive:disable:package-comments
package renderer

import (
	"context"
	"fmt"
	"path"

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

// Type identifies which renderer backend should process the source.
type Type string

const (
	// Helm indicates a Helm chart (Chart.yaml present).
	Helm Type = "helm"
	// Kustomize indicates a kustomization (kustomization.yaml present).
	Kustomize Type = "kustomize"
	// Raw indicates plain YAML manifests with no renderer framework.
	Raw Type = "raw"
)

// DetectRenderer inspects the filesystem at root to determine the renderer
// type based on marker files.
func DetectRenderer(fs billy.Filesystem, root string) Type {
	if fileExists(fs, path.Join(root, "Chart.yaml")) {
		return Helm
	}

	for _, name := range []string{"kustomization.yaml", "kustomization.yml", "Kustomization"} {
		if fileExists(fs, path.Join(root, name)) {
			return Kustomize
		}
	}

	return Raw
}

func fileExists(fs billy.Filesystem, name string) bool {
	info, err := fs.Stat(name)
	return err == nil && !info.IsDir()
}
