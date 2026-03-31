//revive:disable:package-comments
package renderer

import (
	"path/filepath"
	"slices"

	"github.com/go-git/go-billy/v5"

	pb "github.com/katastroma/keleustes"
)

var kustomizeMarkers = []string{"kustomization.yaml", "kustomization.yml", "Kustomization"}

// Detect inspects the filesystem at root to determine the renderer type.
func Detect(fs billy.Filesystem, root string) pb.RendererType {
	if exists(fs, filepath.Join(root, "Chart.yaml")) {
		return pb.RendererType_RENDERER_TYPE_HELM
	}

	if slices.ContainsFunc(kustomizeMarkers, existsAt(fs, root)) {
		return pb.RendererType_RENDERER_TYPE_KUSTOMIZE
	}

	return pb.RendererType_RENDERER_TYPE_PLAIN
}

func exists(fs billy.Filesystem, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}

func existsAt(fs billy.Filesystem, root string) func(string) bool {
	return func(name string) bool {
		return exists(fs, filepath.Join(root, name))
	}
}
