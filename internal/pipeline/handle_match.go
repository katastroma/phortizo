//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/grpc"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/source"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(
	ctx context.Context,
	tenantID string,
	m match.Result,
) {
	ctx, span := tracer.Start(ctx, "pipeline.run", trace.WithAttributes(
		attribute.String("tenant", tenantID),
		attribute.String("watch_target.repo_url", m.Target.RepoURL),
		attribute.String("watch_target.ref", m.Target.Ref),
		attribute.String("watch_target.path", m.Target.Path),
	))
	defer span.End()

	// TODO Retrieve credentials from k8s

	// TODO Parse credentials for type

	// If necessary, exchange credentials for token

	var authMethod transport.AuthMethod

	fs, err := git.Clone(ctx, m.Target.RepoURL, m.Target.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "clone failed", "tenant", tenantID, "error", err)
		return
	}

	rendererType := source.DetectRenderer(fs, m.Target.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err := fmt.Errorf("no renderer configured for type %q", rendererType)
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"renderer lookup failed",
			"tenant", tenantID,
			"renderer", string(rendererType),
			"error", err,
		)
		return
	}

	err = grpc.StreamToRenderer(ctx, tracer, fs, m.Target.Path, rendererAddr)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"streaming to renderer failed",
			"tenant", tenantID,
			"renderer", string(rendererType),
			"error", err,
		)
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", tenantID,
		"renderer", string(rendererType),
	)
}
