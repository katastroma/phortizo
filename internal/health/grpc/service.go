//revive:disable:package-comments
package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/katastroma/phortizo/internal/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const watchInterval = 30 * time.Second

// Service implements the gRPC health check protocol.
type Service struct {
	log *slog.Logger
	healthpb.UnimplementedHealthServer
	checker health.Checker
}

// New returns a gRPC health service.
func New(log *slog.Logger, checker health.Checker) *Service {
	return &Service{log: log, checker: checker}
}

func (s *Service) check(ctx context.Context) healthpb.HealthCheckResponse_ServingStatus {
	if err := s.checker.Check(ctx); err != nil {
		s.log.ErrorContext(ctx, "failed to check upstream health", "error", err)
		return healthpb.HealthCheckResponse_NOT_SERVING
	}
	return healthpb.HealthCheckResponse_SERVING
}

// Check reports service health including GitHub API connectivity.
func (s *Service) Check(ctx context.Context, _ *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: s.check(ctx)}, nil
}

// Watch streams health status changes to the client.
func (s *Service) Watch(_ *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	var last healthpb.HealthCheckResponse_ServingStatus = -1

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	for {
		srvContext := srv.Context()
		current := s.check(srvContext)
		if current != last {
			if err := srv.Send(&healthpb.HealthCheckResponse{Status: current}); err != nil {
				// FIXME No telemetry on failure
				s.log.ErrorContext(srvContext, "failed to send health response", "error", err)
				return err
			}
			last = current
		}

		select {
		case <-srvContext.Done():
			return srvContext.Err()
		case <-ticker.C:
		}
	}
}
