//revive:disable:package-comments
package git

import (
	"context"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Cloner wraps Clone as a type for interface satisfaction.
type Cloner struct{}

// Clone performs a shallow in-memory git clone.
func (Cloner) Clone(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error) {
	return Clone(ctx, url, ref, auth)
}
