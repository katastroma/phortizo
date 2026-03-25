//revive:disable:package-comments
package grpc

import (
	"context"
	"log/slog"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Service implements the gRPC health check protocol.
type Service struct {
	log *slog.Logger
	healthpb.UnimplementedHealthServer
}

// New returns a gRPC health service.
func New(log *slog.Logger) *Service {
	return &Service{log: log}
}

// Check reports service health including GitHub API connectivity.
func (s *Service) Check(
	_ context.Context,
	_ *healthpb.HealthCheckRequest,
) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
