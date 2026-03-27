//revive:disable:package-comments
package tempo

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	"github.com/katastroma/phortizo/internal/tracing"
)

// GetTrace retrieves the full trace and returns all spans grouped by name.
func (q *Tracer) GetTrace(ctx context.Context, traceID string) (tracing.Trace, error) {
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

	return groupSpansByName(resp.Trace), nil
}

func groupSpansByName(trace *tempopb.Trace) tracing.Trace {
	result := make(tracing.Trace)
	for _, batch := range trace.Batches {
		for _, ils := range batch.InstrumentationLibrarySpans {
			for _, span := range ils.Spans {
				result[span.Name] = append(result[span.Name], extractAttributes(span.Attributes))
			}
		}
	}

	return result
}

func extractAttributes(kvs []*commonv1.KeyValue) tracing.Attributes {
	attrs := make(tracing.Attributes, len(kvs))
	for _, kv := range kvs {
		if kv.Value != nil && kv.Value.GetStringValue() != "" {
			attrs[kv.Key] = kv.Value.GetStringValue()
		}
	}

	return attrs
}
