//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/onboarding"
	"github.com/katastroma/phortizo/internal/source"
)

func (r *Runner) fail(
	ctx context.Context,
	span trace.Span,
	namespace, msg string,
	err error,
	attrs ...any,
) {
	args := []any{"tenant", namespace, "error", err}
	args = append(args, attrs...)
	r.log.ErrorContext(ctx, msg, args...)
	span.RecordError(err)
}

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(
	ctx context.Context,
	namespace string,
	m onboarding.WatchTarget,
	replayCount int,
) {
	ctx, span := tracer.Start(ctx, SpanName, trace.WithAttributes(
		attribute.String("tenant", namespace),
		attribute.String("watch_target.name", m.Name),
		attribute.String("watch_target.repo_url", m.RepoURL),
		attribute.String("watch_target.ref", m.Ref),
		attribute.String("watch_target.path", m.Path),
	))
	defer span.End()

	runID := span.SpanContext().TraceID().String()

	if err := configmap.AcquireLease(ctx, r.k8sClient, namespace, m.Name, runID, replayCount); err != nil {
		r.fail(ctx, span, namespace, "lease acquisition failed", err)
		return
	}

	authMethod, err := r.resolveAuth(ctx, namespace, m)
	if err != nil {
		r.fail(ctx, span, namespace, "authentication failed", err)
		return
	}

	fs, err := r.cloner.Clone(ctx, m.RepoURL, m.Ref, authMethod)
	if err != nil {
		r.fail(ctx, span, namespace, "clone failed", err)
		return
	}

	rendererType := source.DetectRenderer(fs, m.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err = fmt.Errorf("no renderer configured for type %q", rendererType)
		r.fail(ctx, span, namespace, "renderer lookup failed", err, "renderer", string(rendererType))
		return
	}

	holds, err := configmap.HoldsLease(ctx, r.k8sClient, namespace, m.Name, runID)
	if err != nil {
		r.fail(ctx, span, namespace, "lease check failed", err)
		return
	}

	if !holds {
		r.log.InfoContext(ctx, "lease lost, abandoning run", "tenant", namespace, "run_id", runID)
		return
	}

	if err = r.renderer.Render(ctx, fs, m.Path, rendererAddr); err != nil {
		r.fail(ctx, span, namespace, "streaming to renderer failed", err, "renderer", string(rendererType))
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", namespace,
		"renderer", string(rendererType),
	)
}
