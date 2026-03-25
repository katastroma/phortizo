//revive:disable:package-comments
package tracequery

import "context"

// Attributes is a set of span attributes keyed by name.
type Attributes map[string]string

// Querier retrieves span attributes from a trace backend.
type Querier interface {
	SpanAttributes(ctx context.Context, traceID string, spanName string) (Attributes, error)
}
