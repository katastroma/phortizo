//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
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
	tracer trace.Tracer,
	namespace string,
	m *source.Target,
	replayCount int,
) {
	ctx, span := tracer.Start(ctx, tracing.WatchTargetSpanName, trace.WithAttributes(
		attribute.String(tracing.WatchTargetNameAttribute, m.Name),
		attribute.String(tracing.WatchTargetRepoURLAttribute, m.RepoURL),
		attribute.String(tracing.WatchTargetRefAttribute, m.Ref),
		attribute.String(tracing.WatchTargetPathAttribute, m.Path),
	))
	defer span.End()

	watchTargetLeaseID := span.SpanContext().SpanID().String()
	store := configmap.NewStore(r.k8sClient, namespace)

	err := lease.Acquire(ctx, store, m.Name, watchTargetLeaseID, replayCount)
	if err != nil {
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

	rendererType := renderer.Detect(fs, m.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err = fmt.Errorf("no renderer configured for type %q", rendererType)
		r.fail(ctx, span, namespace, "renderer lookup failed", err, "renderer", string(rendererType))
		return
	}

	holds, err := lease.HeldBy(ctx, store, m.Name, watchTargetLeaseID)
	if err != nil {
		r.fail(ctx, span, namespace, "lease check failed", err)
		return
	}
	if !holds {
		r.log.InfoContext(ctx, "lease lost, abandoning processing", "tenant", namespace, "watch_target_lease_id", watchTargetLeaseID)
		return
	}

	if err = r.renderer.Stream(ctx, fs, m.Path, rendererAddr); err != nil {
		r.fail(ctx, span, namespace, "streaming to renderer failed", err, "renderer", string(rendererType))
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", namespace,
		"renderer", string(rendererType),
	)
}
