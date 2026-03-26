//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/source"
)

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(
	ctx context.Context,
	namespace string,
	m registration.WatchTarget,
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
		span.RecordError(err)
		r.log.ErrorContext(ctx, "lease acquisition failed", "tenant", namespace, "error", err)
		return
	}

	var authMethod transport.AuthMethod

	if m.CredentialSecret != "" {
		cred, err := r.credentials.Get(ctx, namespace, m.CredentialSecret)
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

	fs, err := r.cloner.Clone(ctx, m.RepoURL, m.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "clone failed", "tenant", namespace, "error", err)
		return
	}

	rendererType := source.DetectRenderer(fs, m.Path)
	rendererAddr, ok := r.renderers[rendererType]
	if !ok {
		err = fmt.Errorf("no renderer configured for type %q", rendererType)
		span.RecordError(err)
		r.log.ErrorContext(ctx, "renderer lookup failed",
			"tenant", namespace,
			"renderer", string(rendererType),
			"error", err,
		)
		return
	}

	holds, err := configmap.HoldsLease(ctx, r.k8sClient, namespace, m.Name, runID)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "lease check failed", "tenant", namespace, "error", err)
		return
	}

	if !holds {
		r.log.InfoContext(ctx, "lease lost, abandoning run", "tenant", namespace, "run_id", runID)
		return
	}

	if err = r.renderer.Render(ctx, fs, m.Path, rendererAddr); err != nil {
		span.RecordError(err)
		r.log.ErrorContext(ctx, "streaming to renderer failed",
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
