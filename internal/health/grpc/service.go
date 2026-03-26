//revive:disable:package-comments
package grpc

import (
	"context"
	"log/slog"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/katastroma/phortizo/internal/health"
)

// Service implements the gRPC health check protocol.
type Service struct {
	log      *slog.Logger
	services []health.Service
	healthpb.UnimplementedHealthServer
}

// New returns a gRPC health service.
func New(log *slog.Logger, services ...health.Service) *Service {
	return &Service{log: log, services: services}
}

// Check reports service health by checking all registered services.
func (s *Service) Check(
	ctx context.Context,
	_ *healthpb.HealthCheckRequest,
) (*healthpb.HealthCheckResponse, error) {
	for _, svc := range s.services {
		if err := svc.Health(ctx); err != nil {
			s.log.ErrorContext(ctx, "health check failed", "service", svc.Name(), "error", err)
			return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_NOT_SERVING}, nil
		}
	}

	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
