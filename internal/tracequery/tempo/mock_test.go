package tempo

import (
	"context"

	"github.com/grafana/tempo/pkg/tempopb"
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
