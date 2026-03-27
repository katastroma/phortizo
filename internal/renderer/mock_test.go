package renderer

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	pb "github.com/katastroma/keleustes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// errorFile errors on Read.
type errorFile struct{ billy.File }

func (f *errorFile) Read(_ []byte) (int, error) { return 0, fmt.Errorf("disk error") }
func (f *errorFile) Close() error               { return nil }

// errorFS wraps a real filesystem but returns an errorFile on Open.
type errorFS struct{ billy.Filesystem }

func (fs *errorFS) Open(_ string) (billy.File, error) {
	return &errorFile{}, nil
}

func (fs *errorFS) Stat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Stat(name)
}

// openErrorFS wraps a real filesystem but errors on Open.
type openErrorFS struct{ billy.Filesystem }

func (fs *openErrorFS) Open(_ string) (billy.File, error) {
	return nil, fmt.Errorf("open denied")
}

func (fs *openErrorFS) Stat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Stat(name)
}

type mockClient struct {
	stream *mockStream
	err    error
}

func (c *mockClient) Render(_ context.Context, _ ...grpc.CallOption) (pb.RendererService_RenderClient, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.stream, nil
}

type mockStream struct {
	sent         []*pb.RenderRequest
	sendErr      error
	closeSendErr error
	grpc.ClientStream
}

func (s *mockStream) Send(req *pb.RenderRequest) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.sent = append(s.sent, req)
	return nil
}

func (s *mockStream) CloseSend() error {
	return s.closeSendErr
}

func (s *mockStream) Recv() (*pb.RenderResponse, error) {
	return nil, io.EOF
}

func (s *mockStream) Header() (metadata.MD, error) { return nil, nil }
func (s *mockStream) Trailer() metadata.MD         { return nil }
func (s *mockStream) Context() context.Context     { return context.Background() }
func (s *mockStream) SendMsg(_ any) error          { return nil }
func (s *mockStream) RecvMsg(_ any) error          { return io.EOF }
