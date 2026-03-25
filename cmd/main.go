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

	"github.com/grafana/tempo/pkg/tempopb"
	pb "github.com/katastroma/naukleros"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gh_apps "github.com/katastroma/phortizo/internal/github/apps"
	grpc_health "github.com/katastroma/phortizo/internal/health/grpc"
	http_health "github.com/katastroma/phortizo/internal/health/http"
	"github.com/katastroma/phortizo/internal/pipeline"
	"github.com/katastroma/phortizo/internal/retriever"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracequery"
	tempoQuerier "github.com/katastroma/phortizo/internal/tracequery/tempo"
	"github.com/katastroma/phortizo/internal/webhook"
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

	gha, err := gh_apps.FromAppParameters(clientID, []byte(privateKeyPEM))
	if err != nil {
		log.Error("Could not create platform GitHub App")
		os.Exit(1)
	}

	// TODO Wire up GitHub client auth bootstrapping:
	// 1. Create ghCred from gha + installationID + gh_api.Service
	// 2. Wrap ghCred in AuthTransport
	// 3. Create gh.Client with that transport
	// 4. Create gh_service from that client
	// 5. Solve JWT refresh for the service used by the credential
	// 6. Pass the authenticated service to health checks
	_ = gha
	_ = installationID

	renderers := map[source.RendererType]string{
		source.Helm:      os.Getenv("RENDERER_HELM"),
		source.Kustomize: os.Getenv("RENDERER_KUSTOMIZE"),
		source.Raw:       os.Getenv("RENDERER_RAW"),
	}

	// TODO Create client connection to k8s for managing
	// - platform GitHub app credentials (client ID, private key, installation ID)
	// - tenant webhook secret
	// - tenant repo credentials
	//   - basic auth
	//   - SSH key
	//   - GitHub token
	// - tenant watch targets

	runner := pipeline.New(log, renderers)

	// HTTP server
	webhookHandler := webhook.New(log, runner)

	mux := http.NewServeMux()
	// TODO Pass authenticated ghService to health check once client auth is wired
	mux.Handle("GET /healthz", http_health.New(log))
	mux.Handle("POST /webhook/{registration_id}", webhookHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	httpServer := &http.Server{Addr: fmt.Sprintf(":%s", httpPort), Handler: mux}

	// gRPC server
	grpcServer := server.New(log)

	healthServer := grpc_health.New(log)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Trace querier — connects to Tempo for replay support
	var traceQuerier tracequery.Querier
	if tempoAddr := os.Getenv("TEMPO_ADDRESS"); tempoAddr != "" {
		tempoConn, err := grpc.NewClient(
			tempoAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Error("connecting to tempo", "address", tempoAddr, "error", err)
			os.Exit(1)
		}
		defer tempoConn.Close()

		qc := tempopb.NewQuerierClient(tempoConn)
		traceQuerier = tempoQuerier.New(qc)
	}

	retriverServer := retriever.New(log, traceQuerier, runner)
	pb.RegisterRetrieverServiceServer(grpcServer, retriverServer)

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
