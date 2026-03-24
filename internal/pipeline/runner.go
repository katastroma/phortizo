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

	"github.com/katastroma/phortizo/internal/auth"
	"github.com/katastroma/phortizo/internal/credential"
	gitclone "github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/source"
)

var tracer = otel.Tracer("pipeline")

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log         *slog.Logger
	credentials credential.Store
	exchanger   auth.TokenExchanger
}

// NewRunner creates a pipeline runner.
func NewRunner(log *slog.Logger, credentials credential.Store, exchanger auth.TokenExchanger) *Runner {
	return &Runner{
		log:         log,
		credentials: credentials,
		exchanger:   exchanger,
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

	r.log.InfoContext(ctx, "source ready",
		"tenant", tenant,
		"renderer", string(rendererType),
	)

	// TODO: stream source to renderer via keleustēs gRPC interface
}

func (r *Runner) resolveCredentials(ctx context.Context, ref string) (*credential.Credential, error) {
	ctx, span := tracer.Start(ctx, "pipeline.resolve_credentials")
	defer span.End()

	cred, err := r.credentials.Get(ctx, ref)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("resolving credentials: %w", err)
	}
	return cred, nil
}

func (r *Runner) resolveAuth(ctx context.Context, cred *credential.Credential) (transport.AuthMethod, error) {
	ctx, span := tracer.Start(ctx, "pipeline.resolve_auth")
	defer span.End()

	method, err := auth.Resolve(ctx, cred, r.exchanger)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("resolving auth: %w", err)
	}
	return method, nil
}

func (r *Runner) clone(ctx context.Context, repoURL, ref string, authMethod transport.AuthMethod) (billy.Filesystem, error) {
	ctx, span := tracer.Start(ctx, "pipeline.clone")
	defer span.End()

	fs, err := gitclone.Clone(ctx, repoURL, ref, authMethod)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("cloning: %w", err)
	}
	return fs, nil
}

func (r *Runner) inspect(ctx context.Context, fs billy.Filesystem, path string) source.RendererType {
	_, span := tracer.Start(ctx, "pipeline.inspect")
	defer span.End()

	return source.DetectRenderer(fs, path)
}
