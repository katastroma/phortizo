//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/source"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Handler processes matched watch targets through the pipeline.
type Handler interface {
	HandleMatch(ctx context.Context, m match.Result)
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

	cred, err := r.vault.Withdraw(ctx, m.Registration.CredentialRef)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "credential resolution failed", "tenant", tenant, "error", err)
		return
	}

	authMethod := cred.Authenticate(ctx)

	fs, err := git.Clone(ctx, m.Target.RepoURL, m.Target.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "clone failed", "tenant", tenant, "error", err)
		return
	}

	rendererType := source.DetectRenderer(fs, m.Target.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err := fmt.Errorf("no renderer configured for type %q", rendererType)
		span.RecordError(err)
		r.log.ErrorContext(ctx, "renderer lookup failed", "tenant", tenant, "renderer", string(rendererType), "error", err)
		return
	}

	if err := r.stream(ctx, fs, m.Target.Path, rendererAddr); err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "streaming to renderer failed", "tenant", tenant, "renderer", string(rendererType), "error", err)
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", tenant,
		"renderer", string(rendererType),
	)
}
