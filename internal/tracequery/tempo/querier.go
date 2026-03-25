//revive:disable:package-comments
package tempo

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	"google.golang.org/grpc"

	"github.com/katastroma/phortizo/internal/tracequery"
)

var _ tracequery.Querier = (*Querier)(nil)

// Querier queries Tempo for span attributes via gRPC.
type Querier struct {
	client tempopb.QuerierClient
}

// New creates a Tempo trace querier from a gRPC connection.
func New(conn *grpc.ClientConn) *Querier {
	return &Querier{client: tempopb.NewQuerierClient(conn)}
}

// SpanAttributes retrieves attributes from the named span within the given
// trace.
func (q *Querier) SpanAttributes(ctx context.Context, traceID string, spanName string) (tracequery.Attributes, error) {
	traceIDBytes, err := hex.DecodeString(traceID)
	if err != nil {
		return nil, fmt.Errorf("decoding trace ID %q: %w", traceID, err)
	}

	resp, err := q.client.FindTraceByID(ctx, &tempopb.TraceByIDRequest{
		TraceID: traceIDBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("querying trace %s: %w", traceID, err)
	}

	if resp.Trace == nil {
		return nil, fmt.Errorf("trace %s not found", traceID)
	}

	return findSpanAttributes(resp.Trace, spanName)
}

func findSpanAttributes(trace *tempopb.Trace, spanName string) (tracequery.Attributes, error) {
	for _, batch := range trace.Batches {
		for _, ils := range batch.InstrumentationLibrarySpans {
			for _, span := range ils.Spans {
				if span.Name == spanName {
					return extractAttributes(span.Attributes), nil
				}
			}
		}
	}
	return nil, fmt.Errorf("span %q not found in trace", spanName)
}

func extractAttributes(kvs []*commonv1.KeyValue) tracequery.Attributes {
	attrs := make(tracequery.Attributes, len(kvs))
	for _, kv := range kvs {
		if kv.Value != nil && kv.Value.GetStringValue() != "" {
			attrs[kv.Key] = kv.Value.GetStringValue()
		}
	}
	return attrs
}
