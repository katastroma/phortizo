//revive:disable:package-comments
package tempo

import (
	"github.com/grafana/tempo/pkg/tempopb"

	"github.com/katastroma/phortizo/internal/tracing"
)

var _ tracing.Tracer = (*Tracer)(nil)

// Tracer queries Tempo for trace data via gRPC
type Tracer struct{ client tempopb.QuerierClient }

// New creates a Tempo trace querier
func New(client tempopb.QuerierClient) *Tracer {
	return &Tracer{client: client}
}

// GetClient gets the querier client of the tracer
func (t *Tracer) GetClient() tempopb.QuerierClient {
	return t.client
}
