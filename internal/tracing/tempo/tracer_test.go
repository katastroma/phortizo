package tempo

import (
	"context"
	"testing"

	"github.com/grafana/tempo/pkg/tempopb"
	commonv1 "github.com/grafana/tempo/pkg/tempopb/common/v1"
	"google.golang.org/grpc"
)

type mockQuerierClient struct {
	resp *tempopb.TraceByIDResponse
	err  error
}

func (m *mockQuerierClient) FindTraceByID(
	_ context.Context, _ *tempopb.TraceByIDRequest, _ ...grpc.CallOption,
) (*tempopb.TraceByIDResponse, error) {
	return m.resp, m.err
}

func (m *mockQuerierClient) SearchRecent(
	_ context.Context, _ *tempopb.SearchRequest, _ ...grpc.CallOption,
) (*tempopb.SearchResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchBlock(
	_ context.Context, _ *tempopb.SearchBlockRequest, _ ...grpc.CallOption,
) (*tempopb.SearchResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchTags(
	_ context.Context, _ *tempopb.SearchTagsRequest, _ ...grpc.CallOption,
) (*tempopb.SearchTagsResponse, error) {
	return nil, nil
}

func (m *mockQuerierClient) SearchTagValues(
	_ context.Context, _ *tempopb.SearchTagValuesRequest, _ ...grpc.CallOption,
) (*tempopb.SearchTagValuesResponse, error) {
	return nil, nil
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
