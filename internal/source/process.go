//revive:disable:package-comments
package source

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/tracing"
)

func fail(
	ctx context.Context,
	log *slog.Logger,
	span trace.Span,
	namespace, msg string,
	err error,
	attrs ...any,
) {
	args := []any{"tenant", namespace, "error", err}
	args = append(args, attrs...)
	log.ErrorContext(ctx, msg, args...)
	span.RecordError(err)
}

// Process executes the source target pipeline steps in order.
func (t *Target) Process(
	ctx context.Context,
	log *slog.Logger,
	tracer trace.Tracer,
	namespace string,
	replayCount int,
	acquireLease lease.AcquireFunc,
	resolveAuth credential.ResolveFunc,
	clone git.CloneFunc,
	verifyLease lease.VerifyFunc,
	stream renderer.StreamFunc,
) {
	ctx, span := tracer.Start(ctx, tracing.SourceTargetSpanName, trace.WithAttributes(
		attribute.String(tracing.SourceTargetNameAttribute, t.Name),
		attribute.String(tracing.SourceTargetRepoURLAttribute, t.RepoURL),
		attribute.String(tracing.SourceTargetRefAttribute, t.Ref),
		attribute.String(tracing.SourceTargetPathAttribute, t.Path),
	))
	defer span.End()

	leaseID := span.SpanContext().SpanID().String()

	if err := acquireLease(ctx, namespace, t.Name, leaseID, replayCount); err != nil {
		fail(ctx, log, span, namespace, "lease acquisition failed", err)
		return
	}

	authMethod, err := resolveAuth(ctx, namespace, t.CredentialSecret)
	if err != nil {
		fail(ctx, log, span, namespace, "authentication failed", err)
		return
	}

	fs, err := clone(ctx, t.RepoURL, t.Ref, authMethod)
	if err != nil {
		fail(ctx, log, span, namespace, "clone failed", err)
		return
	}

	holds, err := verifyLease(ctx, namespace, t.Name, leaseID)
	if err != nil {
		fail(ctx, log, span, namespace, "lease check failed", err)
		return
	}

	if !holds {
		log.InfoContext(ctx, "lease lost, abandoning processing",
			"tenant", namespace,
			"source_target_lease_id", leaseID,
		)
		return
	}

	if err = stream(ctx, fs, t.Path); err != nil {
		fail(ctx, log, span, namespace, "streaming to renderer failed", err)
		return
	}

	log.InfoContext(ctx, "source streamed to renderer", "tenant", namespace)
}
