//revive:disable:package-comments
package pipeline

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/source"
)

func (r *Runner) resolveAuth(
	ctx context.Context,
	namespace string,
	m *source.Target,
) (transport.AuthMethod, error) {
	if m.CredentialSecret == "" {
		return nil, nil
	}

	store := secret.NewStore(r.k8sClient, namespace)
	cred, err := r.credentials.Get(ctx, store, m.CredentialSecret)
	if err != nil {
		return nil, err
	}

	return cred.Authenticate(ctx, r.httpClient)
}
