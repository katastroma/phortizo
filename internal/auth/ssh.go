//revive:disable:package-comments
package auth

import (
	"fmt"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// FromSSHKey returns an AuthMethod for git operations using an SSH key.
func FromSSHKey(pemBytes []byte) (*gitssh.PublicKeys, error) {
	keys, err := gitssh.NewPublicKeys("git", pemBytes, "")
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key: %w", err)
	}

	return keys, nil
}
