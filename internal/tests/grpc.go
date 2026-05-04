//revive:disable:package-comments
package tests

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MockClientConn implements grpc.ClientConnInterface for testing.
type MockClientConn struct {
	grpc.ClientConnInterface
	NewStreamFn func() (grpc.ClientStream, error)
}

// NewStream returns a mock client stream.
func (c *MockClientConn) NewStream(
	_ context.Context, _ *grpc.StreamDesc, _ string, _ ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return c.NewStreamFn()
}

// MockClientStream implements grpc.ClientStream for testing through
// grpc.ClientConnInterface where generated clients call SendMsg.
type MockClientStream struct {
	// SendMsgCount tracks how many messages were sent.
	SendMsgCount int
	// SendMsgErr is returned by SendMsg when set.
	SendMsgErr error
	// Ctx is returned by Context.
	Ctx      context.Context
	recvDone bool
}

// SendMsg records a sent message.
func (s *MockClientStream) SendMsg(_ any) error {
	if s.SendMsgErr != nil {
		return s.SendMsgErr
	}

	s.SendMsgCount++
	return nil
}

// RecvMsg returns nil on the first call (server response) then EOF.
func (s *MockClientStream) RecvMsg(_ any) error {
	if s.recvDone {
		return io.EOF
	}
	s.recvDone = true
	return nil
}

// CloseSend completes the send side.
func (s *MockClientStream) CloseSend() error { return nil }

// Header returns empty metadata.
func (s *MockClientStream) Header() (metadata.MD, error) { return nil, nil }

// Trailer returns empty metadata.
func (s *MockClientStream) Trailer() metadata.MD { return nil }

// Context returns the configured context.
func (s *MockClientStream) Context() context.Context { return s.Ctx }
