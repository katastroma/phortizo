//revive:disable:package-comments
package renderer

import (
	"path"

	"github.com/go-git/go-billy/v5"
)

// Detect inspects the filesystem at root to determine the renderer
// type based on marker files.
func Detect(fs billy.Filesystem, root string) Type {
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
