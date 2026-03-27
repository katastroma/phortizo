//revive:disable:package-comments
package tracing

import (
	"context"
)

// Attributes is a set of span attributes keyed by name.
type Attributes map[string]string

// Trace maps span names to their attributes within a trace.
type Trace map[string][]Attributes

// Tracer retrieves trace data from a trace backend.
type Tracer interface {
	GetTrace(ctx context.Context, traceID string) (Trace, error)
}
