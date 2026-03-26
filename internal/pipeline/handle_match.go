//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/source"
)

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(ctx context.Context, namespace string, m match.Result) {
	ctx, span := tracer.Start(ctx, SpanName, trace.WithAttributes(
		attribute.String("tenant", namespace),
		attribute.String("watch_target.repo_url", m.Target.RepoURL),
		attribute.String("watch_target.ref", m.Target.Ref),
		attribute.String("watch_target.path", m.Target.Path),
	))
	defer span.End()

	var authMethod transport.AuthMethod

	if m.Target.CredentialSecret != "" {
		cred, err := r.credentials.Get(ctx, namespace, m.Target.CredentialSecret)
		if err != nil {
			span.RecordError(err)
			r.log.ErrorContext(ctx, "credential retrieval failed", "tenant", namespace, "error", err)
			return
		}

		authMethod, err = cred.Authenticate(ctx, r.httpClient)
		if err != nil {
			span.RecordError(err)
			r.log.ErrorContext(ctx, "authentication failed", "tenant", namespace, "error", err)
			return
		}
	}

	fs, err := r.cloner.Clone(ctx, m.Target.RepoURL, m.Target.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "clone failed", "tenant", namespace, "error", err)
		return
	}

	rendererType := source.DetectRenderer(fs, m.Target.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err = fmt.Errorf("no renderer configured for type %q", rendererType)
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"renderer lookup failed",
			"tenant", namespace,
			"renderer", string(rendererType),
			"error", err,
		)
		return
	}

	if err = r.renderer.Render(ctx, fs, m.Target.Path, rendererAddr); err != nil {
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"streaming to renderer failed",
			"tenant", namespace,
			"renderer", string(rendererType),
			"error", err,
		)
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", namespace,
		"renderer", string(rendererType),
	)
}
