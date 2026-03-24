//revive:disable:package-comments
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"git.sonicoriginal.software/grpc-foundation/logging"
	foundationotel "git.sonicoriginal.software/grpc-foundation/otel"
	"git.sonicoriginal.software/grpc-foundation/server"

	pb "github.com/katastroma/naukleros"

	"github.com/katastroma/phortizo/internal/auth/resolve"
	"github.com/katastroma/phortizo/internal/github"
	handlers "github.com/katastroma/phortizo/internal/handlers"
	grpc_health "github.com/katastroma/phortizo/internal/health/grpc"
	http_health "github.com/katastroma/phortizo/internal/health/http"
	"github.com/katastroma/phortizo/internal/pipeline"
	regmemory "github.com/katastroma/phortizo/internal/registration/memory"
	"github.com/katastroma/phortizo/internal/source"
	vaultmemory "github.com/katastroma/phortizo/internal/vault/memory"
)

func main() {
	mainCtx := context.Background()
	log := logging.New("phortizo")

	// OTel tracing, metrics, logging
	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		version = "dev"
	}

	providers, err := foundationotel.Init(mainCtx, "phortizo", version)
	if err != nil {
		log.Error("otel init failed", "error", err)
		os.Exit(1)
	}
	defer providers.Shutdown(mainCtx)

	// GitHub App
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

	platformApp := github.NewApp(clientID, []byte(privateKeyPEM))
	ghc := github.NewClientFromAppInstallation(log, platformApp, installationID)

	// Pipeline
	// TODO: store selection from environment config (k8s in prod, memory in dev)
	credentials := vaultmemory.New()

	githubAppFactory := func(clientID string, privateKeyPEM []byte) resolve.TokenExchanger {
		return github.NewApp(clientID, privateKeyPEM)
	}

	renderers := map[source.RendererType]string{
		source.Helm:      os.Getenv("RENDERER_HELM"),
		source.Kustomize: os.Getenv("RENDERER_KUSTOMIZE"),
		source.Raw:       os.Getenv("RENDERER_RAW"),
	}

	runner := pipeline.NewRunner(log, credentials, platformApp, githubAppFactory, renderers)

	// HTTP server
	registrations := regmemory.New()
	webhookHandler := handlers.NewWebhook(log, registrations, runner.HandleMatch)

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
	grpcServer := server.New(log)
	healthpb.RegisterHealthServer(grpcServer, grpc_health.New(log, ghc))
	pb.RegisterRetrieverServiceServer(grpcServer, handlers.NewRetriever(log))

	sigNotifyContext, stop := context.WithCancel(mainCtx)
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
		addr := server.Address()
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("grpc listen error", "error", err)
			stop()
			return
		}
		log.Info("starting grpc server", "address", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", "error", err)
			stop()
		}
	})

	server.HandleGracefulShutdown(sigNotifyContext, stop, log, grpcServer, providers, 10*time.Second)

	shutdownCtx, shutdownCancel := context.WithTimeout(mainCtx, 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown error", "error", err)
	}

	wg.Wait()
}
