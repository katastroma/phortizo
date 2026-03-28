//revive:disable:package-comments
package source

import (
	"context"
	"log/slog"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

// Process executes the watch target pipeline steps in order.
func (t *Target) Process(
	ctx context.Context,
	log *slog.Logger,
	tracer trace.Tracer,
	namespace string,
	replayCount int,
	acquireLease func(ctx context.Context, namespace, name, leaseID string, replayCount int) error,
	resolveAuth func(ctx context.Context, namespace, credentialSecret string) (transport.AuthMethod, error),
	clone func(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error),
	lookupRenderer func(fs billy.Filesystem, path string) (string, error),
	verifyLease func(ctx context.Context, namespace, name, leaseID string) (bool, error),
	stream func(ctx context.Context, fs billy.Filesystem, path, addr string) error,
) {
	ctx, span := tracer.Start(ctx, tracing.WatchTargetSpanName, trace.WithAttributes(
		attribute.String(tracing.WatchTargetNameAttribute, t.Name),
		attribute.String(tracing.WatchTargetRepoURLAttribute, t.RepoURL),
		attribute.String(tracing.WatchTargetRefAttribute, t.Ref),
		attribute.String(tracing.WatchTargetPathAttribute, t.Path),
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

	rendererAddr, err := lookupRenderer(fs, t.Path)
	if err != nil {
		fail(ctx, log, span, namespace, "renderer lookup failed", err)
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
			"watch_target_lease_id", leaseID,
		)
		return
	}

	if err = stream(ctx, fs, t.Path, rendererAddr); err != nil {
		fail(ctx, log, span, namespace, "streaming to renderer failed", err)
		return
	}

	log.InfoContext(ctx, "source streamed to renderer", "tenant", namespace)
}
