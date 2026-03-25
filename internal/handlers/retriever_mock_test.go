package handler

import (
	"context"
	"sync"

	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/tracequery"
)

type stubQuerier struct {
	attrs tracequery.Attributes
	err   error
}

func (q *stubQuerier) SpanAttributes(_ context.Context, _, _ string) (tracequery.Attributes, error) {
	return q.attrs, q.err
}

type captureMatch struct {
	mu      sync.Mutex
	results []match.Result
}

func (c *captureMatch) handle(_ context.Context, m match.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, m)
}
