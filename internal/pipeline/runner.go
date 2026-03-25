//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/katastroma/keleustes"
	"github.com/katastroma/phortizo/internal/auth/resolve"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/render"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/vault"
)

var tracer = otel.Tracer("pipeline")

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log         *slog.Logger
	credentials vault.Withdrawer
	exchanger   resolve.TokenExchanger
	appFactory  resolve.ExchangerFactory
	renderers   map[source.RendererType]string
}

// New creates a pipeline runner.
func New(log *slog.Logger, credentials vault.Withdrawer, exchanger resolve.TokenExchanger, factory resolve.ExchangerFactory, renderers map[source.RendererType]string) *Runner {
	return &Runner{
		log:         log,
		credentials: credentials,
		exchanger:   exchanger,
		appFactory:  factory,
		renderers:   renderers,
	}
}

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(ctx context.Context, m match.Result) {
	ctx, span := tracer.Start(ctx, "pipeline.run", trace.WithAttributes(
		attribute.String("tenant", m.Registration.TenantID),
		attribute.String("watch_target.repo_url", m.Target.RepoURL),
		attribute.String("watch_target.ref", m.Target.Ref),
		attribute.String("watch_target.path", m.Target.Path),
	))
	defer span.End()

	tenant := m.Registration.TenantID

	cred, err := r.resolveCredentials(ctx, m.Registration.CredentialRef)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "credential resolution failed", "tenant", tenant, "error", err)
		return
	}

	authMethod, err := r.resolveAuth(ctx, cred)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "auth resolution failed", "tenant", tenant, "error", err)
		return
	}

	fs, err := r.clone(ctx, m.Target.RepoURL, m.Target.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "clone failed", "tenant", tenant, "error", err)
		return
	}

	rendererType := r.inspect(ctx, fs, m.Target.Path)

	address, ok := r.renderers[rendererType]
	if !ok {
		err := fmt.Errorf("no renderer configured for type %q", rendererType)
		span.RecordError(err)
		r.log.ErrorContext(ctx, "renderer lookup failed", "tenant", tenant, "renderer", string(rendererType), "error", err)
		return
	}

	if err := r.stream(ctx, fs, m.Target.Path, address); err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "streaming to renderer failed", "tenant", tenant, "renderer", string(rendererType), "error", err)
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", tenant,
		"renderer", string(rendererType),
	)
}

func (r *Runner) resolveCredentials(ctx context.Context, ref string) (*vault.Credential, error) {
	ctx, span := tracer.Start(ctx, "pipeline.resolve_credentials")
	defer span.End()

	cred, err := r.credentials.Withdraw(ctx, ref)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("resolving credentials: %w", err)
	}
	return cred, nil
}

func (r *Runner) resolveAuth(ctx context.Context, cred *vault.Credential) (transport.AuthMethod, error) {
	ctx, span := tracer.Start(ctx, "pipeline.resolve_auth")
	defer span.End()

	method, err := resolve.FromCredential(ctx, cred, r.exchanger, r.appFactory)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("resolving auth: %w", err)
	}
	return method, nil
}

func (r *Runner) clone(ctx context.Context, repoURL, ref string, authMethod transport.AuthMethod) (billy.Filesystem, error) {
	ctx, span := tracer.Start(ctx, "pipeline.clone")
	defer span.End()

	fs, err := git.Clone(ctx, repoURL, ref, authMethod)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("cloning: %w", err)
	}
	return fs, nil
}

func (r *Runner) stream(ctx context.Context, fs billy.Filesystem, path, address string) error {
	ctx, span := tracer.Start(ctx, "pipeline.stream")
	defer span.End()

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("connecting to renderer at %s: %w", address, err)
	}
	defer conn.Close()

	client := pb.NewRendererServiceClient(conn)
	if err := render.Stream(ctx, client, fs, path); err != nil {
		span.RecordError(err)
		return fmt.Errorf("streaming to renderer: %w", err)
	}
	return nil
}

func (r *Runner) inspect(ctx context.Context, fs billy.Filesystem, path string) source.RendererType {
	_, span := tracer.Start(ctx, "pipeline.inspect")
	defer span.End()

	return source.DetectRenderer(fs, path)
}
