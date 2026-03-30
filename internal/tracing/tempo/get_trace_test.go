//revive:disable:package-comments
package tempo

import (
	"fmt"
	"testing"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	tracev1 "github.com/grafana/tempo/pkg/tempopb/trace/v1"
)

func TestGetTrace(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{
		Trace: &tempopb.Trace{
			Batches: []*tracev1.ResourceSpans{{
				InstrumentationLibrarySpans: []*tracev1.InstrumentationLibrarySpans{{
					Spans: []*tracev1.Span{
						{
							Name: "event",
							Attributes: []*commonv1.KeyValue{
								{Key: "tenant", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "tenant-a"}}},
								{Key: "event.type", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "webhook"}}},
							},
						},
						{
							Name: "source_target",
							Attributes: []*commonv1.KeyValue{
								{Key: "source_target.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "wt-1"}}},
							},
						},
						{
							Name: "source_target",
							Attributes: []*commonv1.KeyValue{
								{Key: "source_target.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "wt-2"}}},
							},
						},
					},
				}},
			}},
		},
	}}

	q := &Tracer{client}
	trace, err := q.GetTrace(t.Context(), "abcdef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := trace["event"]
	if len(events) != 1 {
		t.Fatalf("expected 1 event span, got %d", len(events))
	}

	if events[0]["tenant"] != "tenant-a" {
		t.Errorf("tenant = %q, want %q", events[0]["tenant"], "tenant-a")
	}

	targets := trace["source_target"]
	if len(targets) != 2 {
		t.Fatalf("expected 2 source_target spans, got %d", len(targets))
	}

	if targets[0]["source_target.name"] != "wt-1" {
		t.Errorf("first target = %q, want %q", targets[0]["source_target.name"], "wt-1")
	}

	if targets[1]["source_target.name"] != "wt-2" {
		t.Errorf("second target = %q, want %q", targets[1]["source_target.name"], "wt-2")
	}
}

func TestGetTrace_InvalidTraceID(t *testing.T) {
	q := &Tracer{client: &mockQuerierClient{}}

	if _, err := q.GetTrace(t.Context(), "not-hex"); err == nil {
		t.Fatal("expected error for invalid trace ID")
	}
}

func TestGetTrace_QueryError(t *testing.T) {
	client := &mockQuerierClient{err: fmt.Errorf("tempo unavailable")}
	q := &Tracer{client}

	if _, err := q.GetTrace(t.Context(), "abcdef1234567890"); err == nil {
		t.Fatal("expected error for query failure")
	}
}

func TestGetTrace_TraceNotFound(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{Trace: nil}}
	q := &Tracer{client}

	if _, err := q.GetTrace(t.Context(), "abcdef1234567890"); err == nil {
		t.Fatal("expected error for nil trace")
	}
}
