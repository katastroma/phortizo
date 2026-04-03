//revive:disable:package-comments
package source

import (
	"context"
	"errors"
	"fmt"
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

type baggageMemberBuilder struct {
	members []baggage.Member
	errs    []error
	span    trace.Span
}

func (b *baggageMemberBuilder) add(key, value string) {
	m, err := baggage.NewMemberRaw(key, value)
	if err != nil {
		b.span.RecordError(err)
		b.errs = append(b.errs, fmt.Errorf("baggage failed for %v [%v]: %w", key, value, err))
		return
	}
	b.members = append(b.members, m)
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
) error {
	log = log.With("name", t.Name, "repo", t.RepoURL, "ref", t.Ref, "path", t.Path)
	log.InfoContext(ctx, "processing source target")
	ctx, span := tracer.Start(ctx, tracing.SourceTargetSpanName, trace.WithAttributes(
		attribute.String(tracing.SourceTargetNameAttribute, t.Name),
		attribute.String(tracing.SourceTargetRepoURLAttribute, t.RepoURL),
		attribute.String(tracing.SourceTargetRefAttribute, t.Ref),
		attribute.String(tracing.SourceTargetPathAttribute, t.Path),
	))
	defer span.End()

	b := &baggageMemberBuilder{span: span}
	b.add(tracing.TenantAttribute, namespace)
	b.add(tracing.SourceTargetNameAttribute, t.Name)
	b.add(tracing.SourceTargetRepoURLAttribute, t.RepoURL)
	b.add(tracing.SourceTargetRefAttribute, t.Ref)
	b.add(tracing.SourceTargetPathAttribute, t.Path)

	if len(b.errs) > 0 {
		return errors.Join(b.errs...)
	}

	bag, err := baggage.New(b.members...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("baggage creation failed: %w", err)
	}
	ctx = baggage.ContextWithBaggage(ctx, bag)

	log.DebugContext(ctx, "acquiring lease")
	leaseID := span.SpanContext().SpanID().String()
	if err = acquireLease(ctx, namespace, t.Name, leaseID, replayCount); err != nil {
		span.RecordError(err)
		return fmt.Errorf("lease acquisition failed: %w", err)
	}
	log.DebugContext(ctx, "lease acquired")

	log.DebugContext(ctx, "resolving credentials")
	authMethod, err := resolveAuth(ctx, namespace, t.CredentialSecret)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("authentication failed: %w", err)
	}
	log.DebugContext(ctx, "credentials resolved")

	log.DebugContext(ctx, "cloning source")
	fs, err := clone(ctx, t.RepoURL, t.Ref, authMethod)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("clone failed: %w", err)
	}
	log.DebugContext(ctx, "source cloned")

	log.DebugContext(ctx, "detecting renderer")
	rendererType := renderer.Detect(fs, t.Path)
	log = log.With("renderer", rendererType.String())
	log.DebugContext(ctx, "renderer detected")

	log.DebugContext(ctx, "verifying lease")
	holds, err := verifyLease(ctx, namespace, t.Name, leaseID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("lease check failed: %w", err)
	} else if !holds {
		err = fmt.Errorf("lease lost: %v", leaseID)
		span.RecordError(err)
		return err
	}
	log.DebugContext(ctx, "lease verified")

	log.DebugContext(ctx, "streaming to renderer")
	if err = stream(ctx, fs, t.Path, rendererType); err != nil {
		span.RecordError(err)
		return fmt.Errorf("streaming to renderer failed: %w", err)
	}

	log.InfoContext(ctx, "source streamed to renderer")
	return nil
}
