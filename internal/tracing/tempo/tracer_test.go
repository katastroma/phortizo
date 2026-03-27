package tempo

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	tracev1 "github.com/grafana/tempo/pkg/tempopb/trace/v1"
	"google.golang.org/grpc"
)

type mockQuerierClient struct {
	resp *tempopb.TraceByIDResponse
	err  error
}

func (m *mockQuerierClient) FindTraceByID(_ context.Context, _ *tempopb.TraceByIDRequest, _ ...grpc.CallOption) (*tempopb.TraceByIDResponse, error) {
	return m.resp, m.err
}

func (m *mockQuerierClient) SearchRecent(_ context.Context, _ *tempopb.SearchRequest, _ ...grpc.CallOption) (*tempopb.SearchResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchBlock(_ context.Context, _ *tempopb.SearchBlockRequest, _ ...grpc.CallOption) (*tempopb.SearchResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchTags(_ context.Context, _ *tempopb.SearchTagsRequest, _ ...grpc.CallOption) (*tempopb.SearchTagsResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchTagValues(_ context.Context, _ *tempopb.SearchTagValuesRequest, _ ...grpc.CallOption) (*tempopb.SearchTagValuesResponse, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	q := New(&mockQuerierClient{})
	if q.client == nil {
		t.Fatal("expected non-nil client")
	}
}

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
							Name: "watch_target",
							Attributes: []*commonv1.KeyValue{
								{Key: "watch_target.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "wt-1"}}},
							},
						},
						{
							Name: "watch_target",
							Attributes: []*commonv1.KeyValue{
								{Key: "watch_target.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "wt-2"}}},
							},
						},
					},
				}},
			}},
		},
	}}

	q := &Tracer{client: client}
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

	targets := trace["watch_target"]
	if len(targets) != 2 {
		t.Fatalf("expected 2 watch_target spans, got %d", len(targets))
	}

	if targets[0]["watch_target.name"] != "wt-1" {
		t.Errorf("first target = %q, want %q", targets[0]["watch_target.name"], "wt-1")
	}

	if targets[1]["watch_target.name"] != "wt-2" {
		t.Errorf("second target = %q, want %q", targets[1]["watch_target.name"], "wt-2")
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
	q := &Tracer{client: client}

	if _, err := q.GetTrace(t.Context(), "abcdef1234567890"); err == nil {
		t.Fatal("expected error for query failure")
	}
}

func TestGetTrace_TraceNotFound(t *testing.T) {
	client := &mockQuerierClient{resp: &tempopb.TraceByIDResponse{Trace: nil}}
	q := &Tracer{client: client}

	if _, err := q.GetTrace(t.Context(), "abcdef1234567890"); err == nil {
		t.Fatal("expected error for nil trace")
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
