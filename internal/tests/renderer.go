//revive:disable:package-comments
package tests

import (
	"context"

	pb "github.com/katastroma/keleustes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MockRendererClient implements pb.RendererServiceClient for testing.
type MockRendererClient struct {
	// Stream is returned by Render on success.
	Stream *MockRenderStream
	// Err is returned by Render when set.
	Err error
}

// Render returns the configured mock stream or error.
func (c *MockRendererClient) Render(_ context.Context, _ ...grpc.CallOption) (pb.RendererService_RenderClient, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.Stream, nil
}

// RenderStream is a no-op stub to satisfy RendererServiceClient.
func (c *MockRendererClient) RenderStream(_ context.Context, _ ...grpc.CallOption) (pb.RendererService_RenderStreamClient, error) {
	return nil, nil
}

// MockRenderStream implements pb.RendererService_RenderClient for testing.
type MockRenderStream struct {
	// Sent collects all messages passed to Send.
	Sent []*pb.RenderRequest
	// SendErr is returned by Send when set.
	SendErr error
	// CloseSendErr is returned by CloseSend when set.
	CloseSendErr error
	// CloseAndRecvErr is returned by CloseAndRecv when set.
	CloseAndRecvErr error
	// Ctx is returned by Context.
	Ctx context.Context
	grpc.ClientStream
}

// Send records a sent message or returns the configured error.
func (s *MockRenderStream) Send(req *pb.RenderRequest) error {
	if s.SendErr != nil {
		return s.SendErr
	}
	s.Sent = append(s.Sent, req)
	return nil
}

// CloseSend signals the server that the client is done sending.
func (s *MockRenderStream) CloseSend() error {
	return s.CloseSendErr
}

// Header returns empty metadata.
func (s *MockRenderStream) Header() (metadata.MD, error) { return nil, nil }

// Trailer returns empty metadata.
func (s *MockRenderStream) Trailer() metadata.MD { return nil }

// Context returns the configured context.
func (s *MockRenderStream) Context() context.Context { return s.Ctx }

// CloseAndRecv closes the stream and returns the configured error.
func (s *MockRenderStream) CloseAndRecv() (*pb.RenderResponse, error) {
	return &pb.RenderResponse{}, s.CloseAndRecvErr
}

// Recv is a no-op.
func (s *MockRenderStream) Recv() (*pb.RenderResponse, error) { return nil, nil }

// SendMsg is a no-op.
func (s *MockRenderStream) SendMsg(_ any) error { return nil }

// RecvMsg is a no-op.
func (s *MockRenderStream) RecvMsg(_ any) error { return nil }
