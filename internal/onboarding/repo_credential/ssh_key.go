//revive:disable:package-comments
package credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/katastroma/phortizo/internal/auth"
)

// TypeSSHKey is the Secret type for SSH key credentials.
const TypeSSHKey = "ssh-key"

// SSHKey authenticates with an SSH private key.
type SSHKey struct {
	privateKey []byte
}

// NewSSHKey returns an SSHKey credential.
func NewSSHKey(privateKey []byte) *SSHKey {
	return &SSHKey{privateKey: privateKey}
}

// SSHKeyFromSecret deserializes an SSHKey from Secret data.
func SSHKeyFromSecret(data map[string][]byte) (auth.Credential, error) {
	pk, ok := data["private-key"]
	if !ok {
		return nil, fmt.Errorf("missing key %q", "private-key")
	}

	return NewSSHKey(pk), nil
}

// MarshalSecret serializes the credential to Secret data.
func (c *SSHKey) MarshalSecret() map[string][]byte {
	return map[string][]byte{
		"type":        []byte(TypeSSHKey),
		"private-key": c.privateKey,
	}
}

// Authenticate returns an SSH PublicKeys transport.
func (c *SSHKey) Authenticate(context.Context, *http.Client) (transport.AuthMethod, error) {
	return gitssh.NewPublicKeys("git", c.privateKey, "")
}
