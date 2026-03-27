//revive:disable:package-comments
package git

import (
	"context"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Cloner performs a git clone and returns the worktree filesystem
type Cloner interface {
	Clone(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error)
}

// Client client to perform git operations
type Client struct{}

// Clone performs a shallow in-memory git clone
func (Client) Clone(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error) {
	return Clone(ctx, url, ref, auth)
}
