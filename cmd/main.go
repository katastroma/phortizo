//revive:disable:package-comments
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/katastroma/keleusma/broker"
	"github.com/katastroma/kytos/store"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/github"
	handlers "github.com/katastroma/phortizo/internal/handlers"
	grpc_health "github.com/katastroma/phortizo/internal/health/grpc"
	http_health "github.com/katastroma/phortizo/internal/health/http"
	"github.com/katastroma/phortizo/internal/pipeline"
	"github.com/katastroma/phortizo/internal/registration"
)

func main() {
	mainCtx := context.Background()
	log := slog.Default()

	clientID := os.Getenv("GITHUB_APP_CLIENT_ID")
	if clientID == "" {
		log.Error("GITHUB_APP_CLIENT_ID is required")
		os.Exit(1)
	}

	installationID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_INSTALLATION_ID"), 10, 64)
	if err != nil {
		log.Error("GITHUB_APP_INSTALLATION_ID is required and must be an integer", "error", err)
		os.Exit(1)
	}

	privateKeyPEM := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if privateKeyPEM == "" {
		log.Error("GITHUB_APP_PRIVATE_KEY is required")
		os.Exit(1)
	}

	gha := github.NewApp(clientID, []byte(privateKeyPEM))
	ghc := github.NewClientFromAppInstallation(log, gha, installationID)

	// Pipeline dependencies — injected at runtime.
	// TODO: select concrete implementations from environment config.
	var (
		storage     store.Store
		publisher   broker.Publisher
		credentials credential.Store
	)

	runner := pipeline.NewRunner(log, credentials, gha, storage, publisher)

	registrationStore := registration.NewMemoryStore()
	webhookHandler := handlers.NewWebhook(log, registrationStore, runner.HandleMatch)

	// HTTP server
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", http_health.New(log, ghc))
	mux.Handle("POST /webhook/{id}", webhookHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: mux,
	}

	// gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	grpcServer := grpc.NewServer()
	healthpb.RegisterHealthServer(grpcServer, grpc_health.New(log, ghc))

	sigNotifyContext, stop := signal.NotifyContext(mainCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	wg.Go(func() {
		log.Info("starting http server", "port", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", "error", err)
			stop()
		}
	})

	wg.Go(func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Error("grpc listen error", "error", err)
			stop()
			return
		}
		log.Info("starting grpc server", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", "error", err)
			stop()
		}
	})

	<-sigNotifyContext.Done()

	log.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(mainCtx, 10*time.Second)
	defer shutdownCancel()

	var shutdownWg sync.WaitGroup

	shutdownWg.Go(func() {
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-shutdownCtx.Done():
			grpcServer.Stop()
		}
	})

	shutdownWg.Go(func() {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("http shutdown error", "error", err)
		}
	})

	shutdownWg.Wait()
	wg.Wait()
}
