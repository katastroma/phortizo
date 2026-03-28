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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/go-github/v84/github"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/credential/provider"
	"github.com/katastroma/phortizo/internal/git"
	gh_apps "github.com/katastroma/phortizo/internal/github/apps"
	gh_service "github.com/katastroma/phortizo/internal/github/service"
	gh_transport "github.com/katastroma/phortizo/internal/github/transport"
	grpc_health "github.com/katastroma/phortizo/internal/health/grpc"
	http_health "github.com/katastroma/phortizo/internal/health/http"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/retriever"
	tempoTracer "github.com/katastroma/phortizo/internal/tracing/tempo"
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
		log.Error("could not create platform GitHub App", "error", err)
		os.Exit(1)
	}

	// Authenticated GitHub client for platform operations (health checks)
	ghTransport := gh_transport.NewInstallationTokenAuth(gha, installationID, nil)
	ghClient := github.NewClient(&http.Client{Transport: ghTransport})
	ghService := gh_service.New(ghClient)

	// Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Info("not running in cluster, falling back to kubeconfig")
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		k8sConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil).ClientConfig()
		if err != nil {
			log.Error("could not create k8s client config", "error", err)
			os.Exit(1)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Error("could not create k8s client", "error", err)
		os.Exit(1)
	}

	// Credential reader
	credentialReader := provider.NewReader(gha)

	// Renderer addresses
	renderers := map[renderer.Type]string{
		renderer.Helm:      os.Getenv("RENDERER_HELM"),
		renderer.Kustomize: os.Getenv("RENDERER_KUSTOMIZE"),
		renderer.Raw:       os.Getenv("RENDERER_RAW"),
	}

	// Lease configuration
	leaseStaleAfter := 10 * time.Minute
	if raw := os.Getenv("LEASE_STALE_AFTER"); raw != "" {
		leaseStaleAfter, err = time.ParseDuration(raw)
		if err != nil {
			log.Error("LEASE_STALE_AFTER must be a valid duration", "error", err)
			os.Exit(1)
		}
	}

	maxReplayAttempts := 3
	if raw := os.Getenv("MAX_REPLAY_ATTEMPTS"); raw != "" {
		maxReplayAttempts, err = strconv.Atoi(raw)
		if err != nil {
			log.Error("MAX_REPLAY_ATTEMPTS must be an integer", "error", err)
			os.Exit(1)
		}
	}

	// Pipeline step closures
	newConfigMapStore := func(ns string) object.Store[string] {
		return configmap.NewStore(k8sClient, ns)
	}
	newSecretStore := func(ns string) object.Store[[]byte] {
		return secret.NewStore(k8sClient, ns)
	}

	acquireLease := lease.AcquireFunc(newConfigMapStore)
	resolveAuth := credential.ResolveFunc(credentialReader, http.DefaultClient, newSecretStore)
	gitClient := git.Client{}
	lookupRenderer := renderer.LookupFunc(renderers)
	verifyLease := lease.VerifyFunc(newConfigMapStore)

	// Renderer gRPC connections
	rendererConns := make(map[string]grpc.ClientConnInterface, len(renderers))
	for rendererType, addr := range renderers {
		if addr == "" {
			continue
		}

		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error("connecting to renderer", "type", string(rendererType), "address", addr, "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		rendererConns[addr] = conn
	}

	streamToRenderer := renderer.StreamFunc(rendererConns)

	// HTTP server
	webhookHandler := webhook.New(
		log, k8sClient,
		acquireLease, resolveAuth, gitClient.Clone, lookupRenderer,
		verifyLease, streamToRenderer,
	)

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", http_health.New(log, ghService))
	mux.Handle("POST /webhook/{namespace}", webhookHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	httpServer := &http.Server{Addr: fmt.Sprintf(":%s", httpPort), Handler: mux}

	// gRPC server
	grpcServer := server.New(log)

	healthServer := grpc_health.New(log, ghService)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Tracer — connects to Tempo for replay support
	tempoAddr := os.Getenv("TEMPO_ADDRESS")
	if tempoAddr == "" {
		log.Error("TEMPO_ADDRESS is required")
		os.Exit(1)
	}

	tempoConn, err := grpc.NewClient(tempoAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("connecting to tempo", "address", tempoAddr, "error", err)
		os.Exit(1)
	}
	defer tempoConn.Close()

	qc := tempopb.NewQuerierClient(tempoConn)
	tracer := tempoTracer.New(qc)

	retrieverServer := retriever.New(
		log, tracer, k8sClient,
		leaseStaleAfter, maxReplayAttempts,
		acquireLease, resolveAuth, gitClient.Clone, lookupRenderer,
		verifyLease, streamToRenderer,
	)
	pb.RegisterRetrieverServiceServer(grpcServer, retrieverServer)

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
