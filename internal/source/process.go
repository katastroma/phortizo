//revive:disable:package-comments
package source

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
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
	msg string,
	err error,
	attrs ...any,
) {
	args := []any{"error", err}
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

	tenantMember, err := baggage.NewMemberRaw(tracing.TenantAttribute, namespace)
	if err != nil {
		fail(ctx, log, span, "baggage member failed", err, tracing.TenantAttribute, namespace)
		return
	}

	nameMember, err := baggage.NewMemberRaw(tracing.SourceTargetNameAttribute, t.Name)
	if err != nil {
		fail(ctx, log, span, "baggage member failed", err, tracing.SourceTargetNameAttribute, t.Name)
		return
	}

	repoMember, err := baggage.NewMemberRaw(tracing.SourceTargetRepoURLAttribute, t.RepoURL)
	if err != nil {
		fail(ctx, log, span, "baggage member failed", err, tracing.SourceTargetRepoURLAttribute, t.RepoURL)
		return
	}

	refMember, err := baggage.NewMemberRaw(tracing.SourceTargetRefAttribute, t.Ref)
	if err != nil {
		fail(ctx, log, span, "baggage member failed", err, tracing.SourceTargetRefAttribute, t.Ref)
		return
	}

	pathMember, err := baggage.NewMemberRaw(tracing.SourceTargetPathAttribute, t.Path)
	if err != nil {
		fail(ctx, log, span, "baggage member failed", err, tracing.SourceTargetPathAttribute, t.Path)
		return
	}

	bag, err := baggage.New(tenantMember, nameMember, repoMember, refMember, pathMember)
	if err != nil {
		fail(ctx, log, span, "baggage creation failed", err)
		return
	}

	ctx = baggage.ContextWithBaggage(ctx, bag)

	leaseID := span.SpanContext().SpanID().String()

	log.DebugContext(ctx, "acquiring lease")
	if err = acquireLease(ctx, namespace, t.Name, leaseID, replayCount); err != nil {
		fail(ctx, log, span, "lease acquisition failed", err)
		return
	}
	log.DebugContext(ctx, "lease acquired")

	log.DebugContext(ctx, "resolving credentials")
	authMethod, err := resolveAuth(ctx, namespace, t.CredentialSecret)
	if err != nil {
		fail(ctx, log, span, "authentication failed", err)
		return
	}
	log.DebugContext(ctx, "credentials resolved")

	log.DebugContext(ctx, "cloning source")
	fs, err := clone(ctx, t.RepoURL, t.Ref, authMethod)
	if err != nil {
		fail(ctx, log, span, "clone failed", err)
		return
	}
	log.DebugContext(ctx, "source cloned")

	rendererType := renderer.Detect(fs, t.Path)
	log.InfoContext(ctx, "renderer detected", "renderer", rendererType.String())

	log.DebugContext(ctx, "verifying lease")
	holds, err := verifyLease(ctx, namespace, t.Name, leaseID)
	if err != nil {
		fail(ctx, log, span, "lease check failed", err)
		return
	}
	log.DebugContext(ctx, "lease verified")

	if !holds {
		log.InfoContext(ctx, "lease lost, abandoning processing",
			"source_target_lease_id", leaseID,
		)
		return
	}

	log.DebugContext(ctx, "streaming to renderer")
	if err = stream(ctx, fs, t.Path, rendererType); err != nil {
		fail(ctx, log, span, "streaming to renderer failed", err)
		return
	}
	log.InfoContext(ctx, "source streamed to renderer")
}
