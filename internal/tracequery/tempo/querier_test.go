package tempo

import (
	"fmt"
	"testing"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	tracev1 "github.com/grafana/tempo/pkg/tempopb/trace/v1"
)

func TestSpanAttributes(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{
		Trace: &tempopb.Trace{
			Batches: []*tracev1.ResourceSpans{{
				InstrumentationLibrarySpans: []*tracev1.InstrumentationLibrarySpans{{
					Spans: []*tracev1.Span{{
						Name: "pipeline.run",
						Attributes: []*commonv1.KeyValue{
							{Key: "tenant", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "acme"}}},
						},
					}},
				}},
			}},
		},
	}}

	q := &Querier{client: client}
	attrs, err := q.SpanAttributes(t.Context(), "abcdef1234567890", "pipeline.run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attrs["tenant"] != "acme" {
		t.Errorf("tenant = %q, want %q", attrs["tenant"], "acme")
	}
}

func TestSpanAttributes_InvalidTraceID(t *testing.T) {
	q := &Querier{client: &mockQuerierClient{}}

	if _, err := q.SpanAttributes(t.Context(), "not-hex", "pipeline.run"); err == nil {
		t.Fatal("expected error for invalid trace ID")
	}
}

func TestSpanAttributes_QueryError(t *testing.T) {
	client := &mockQuerierClient{err: fmt.Errorf("tempo unavailable")}
	q := &Querier{client: client}

	if _, err := q.SpanAttributes(t.Context(), "abcdef1234567890", "pipeline.run"); err == nil {
		t.Fatal("expected error for query failure")
	}
}

func TestSpanAttributes_TraceNotFound(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{Trace: nil}}
	q := &Querier{client: client}

	if _, err := q.SpanAttributes(t.Context(), "abcdef1234567890", "pipeline.run"); err == nil {
		t.Fatal("expected error for nil trace")
	}
}

func TestSpanAttributes_SpanNotFound(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{
		Trace: &tempopb.Trace{
			Batches: []*tracev1.ResourceSpans{{
				InstrumentationLibrarySpans: []*tracev1.InstrumentationLibrarySpans{{
					Spans: []*tracev1.Span{{Name: "other.span"}},
				}},
			}},
		},
	}}

	q := &Querier{client: client}
	if _, err := q.SpanAttributes(t.Context(), "abcdef1234567890", "pipeline.run"); err == nil {
		t.Fatal("expected error for missing span")
	}
}

func TestExtractAttributes_SkipsNonString(t *testing.T) {
	kvs := []*commonv1.KeyValue{
		{Key: "good", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "value"}}},
		{Key: "empty", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: ""}}},
		{Key: "nil_value", Value: nil},
	}

	attrs := extractAttributes(kvs)
	if attrs["good"] != "value" {
		t.Errorf("good = %q, want %q", attrs["good"], "value")
	}
	if _, ok := attrs["empty"]; ok {
		t.Error("expected empty string attribute to be skipped")
	}
	if _, ok := attrs["nil_value"]; ok {
		t.Error("expected nil value attribute to be skipped")
	}
}
