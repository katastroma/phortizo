//revive:disable:package-comments
package vault

import (
	"context"

	"github.com/katastroma/phortizo/internal/auth"
)

// Withdrawer retrieves credentials for tenants.
type Withdrawer interface {
	Withdraw(ctx context.Context, ref string) (auth.Credential, error)
}
