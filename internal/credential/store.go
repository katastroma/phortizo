//revive:disable:package-comments
package credential

import "context"

// Type identifies the authentication method.
type Type string

const (
	// AppInstallation authenticates using the platform's GitHub App.
	AppInstallation Type = "app-installation"
	// TenantApp authenticates using the tenant's own GitHub App.
	TenantApp Type = "tenant-app"
	// Token authenticates using a GitHub personal or fine-grained token.
	Token Type = "token"
	// BasicAuth authenticates using a username and password.
	BasicAuth Type = "basic"
	// SSH authenticates using an SSH private key.
	SSH Type = "ssh"
)

// Credential holds tenant authentication details for cloning repositories.
type Credential struct {
	Type Type

	// AppInstallation / TenantApp
	InstallationID int64

	// TenantApp
	ClientID      string
	PrivateKeyPEM []byte

	// Token
	Token string

	// BasicAuth
	Username string
	Password string

	// SSH
	SSHKeyPEM []byte
}

// Store retrieves credentials for tenants.
type Store interface {
	Get(ctx context.Context, ref string) (*Credential, error)
}
