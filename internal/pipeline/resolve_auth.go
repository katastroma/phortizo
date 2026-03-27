//revive:disable:package-comments
package pipeline

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/katastroma/phortizo/internal/source"
)

func (r *Runner) resolveAuth(
	ctx context.Context,
	namespace string,
	m source.WatchTarget,
) (transport.AuthMethod, error) {
	if m.CredentialSecret == "" {
		return nil, nil
	}

	cred, err := r.credentials.Get(ctx, namespace, m.CredentialSecret)
	if err != nil {
		return nil, err
	}

	return cred.Authenticate(ctx, r.httpClient)
}
