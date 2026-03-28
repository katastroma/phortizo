//revive:disable:package-comments
package renderer

import (
	"fmt"

	"github.com/go-git/go-billy/v5"
)

// LookupFunc returns a closure that detects the renderer type from the
// filesystem and resolves its service address.
func LookupFunc(
	renderers map[Type]string,
) func(fs billy.Filesystem, path string) (string, error) {
	return func(fs billy.Filesystem, path string) (string, error) {
		rendererType := Detect(fs, path)

		addr, ok := renderers[rendererType]
		if !ok {
			return "", fmt.Errorf("no renderer configured for type %q", rendererType)
		}

		return addr, nil
	}
}
