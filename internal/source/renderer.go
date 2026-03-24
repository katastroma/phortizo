//revive:disable:package-comments
package source

import (
	"path"

	"github.com/go-git/go-billy/v5"
)

// RendererType identifies which renderer backend should process the source.
type RendererType string

const (
	// Helm indicates a Helm chart (Chart.yaml present).
	Helm RendererType = "helm"
	// Kustomize indicates a kustomization (kustomization.yaml present).
	Kustomize RendererType = "kustomize"
	// Raw indicates plain YAML manifests with no renderer framework.
	Raw RendererType = "raw"
)

// DetectRenderer inspects the filesystem at root to determine the renderer
// type based on marker files.
func DetectRenderer(fs billy.Filesystem, root string) RendererType {
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
