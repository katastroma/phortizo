//revive:disable:package-comments
package retriever

import (
	"context"

	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/tracequery"
)

type stubQuerier struct {
	attrs tracequery.Attributes
	err   error
}

func (q *stubQuerier) SpanAttributes(_ context.Context, _, _ string) (tracequery.Attributes, error) {
	return q.attrs, q.err
}

func testRegistration() *registration.Record {
	return &registration.Record{
		ID:            "reg-1",
		TenantID:      "acme",
		Secret:        []byte("test-secret"),
		CredentialRef: "cred-1",
		WatchTargets: []registration.WatchTarget{
			{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		},
	}
}
