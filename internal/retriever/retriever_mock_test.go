//revive:disable:package-comments
package retriever

import (
	"context"

	"github.com/katastroma/phortizo/internal/tracequery"
)

type stubQuerier struct {
	attrs tracequery.Attributes
	err   error
}

func (q *stubQuerier) SpanAttributes(_ context.Context, _, _ string) (tracequery.Attributes, error) {
	return q.attrs, q.err
}
