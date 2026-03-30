//revive:disable:package-comments
package tests

import (
	"context"
	"io"

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

// MockRenderStream implements pb.RendererService_RenderClient for testing.
type MockRenderStream struct {
	// Sent collects all messages passed to Send.
	Sent []*pb.RenderRequest
	// SendErr is returned by Send when set.
	SendErr error
	// CloseSendErr is returned by CloseSend when set.
	CloseSendErr error
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

// CloseSend returns the configured error.
func (s *MockRenderStream) CloseSend() error {
	return s.CloseSendErr
}

// Recv signals end of stream.
func (s *MockRenderStream) Recv() (*pb.RenderResponse, error) {
	return nil, io.EOF
}

// Header returns empty metadata.
func (s *MockRenderStream) Header() (metadata.MD, error) { return nil, nil }

// Trailer returns empty metadata.
func (s *MockRenderStream) Trailer() metadata.MD { return nil }

// Context returns a background context.
func (s *MockRenderStream) Context() context.Context { return context.Background() }

// SendMsg is a no-op satisfying grpc.ClientStream.
func (s *MockRenderStream) SendMsg(_ any) error { return nil }

// RecvMsg signals end of stream.
func (s *MockRenderStream) RecvMsg(_ any) error { return io.EOF }
