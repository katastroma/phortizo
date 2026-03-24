//revive:disable:package-comments
package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Clone performs a shallow in-memory git clone and returns the worktree
// filesystem.
func Clone(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error) {
	fs := memfs.New()
	_, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		URL:           url,
		Auth:          auth,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(ref),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("cloning %s at %s: %w", url, ref, err)
	}

	return fs, nil
}
