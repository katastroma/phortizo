//revive:disable:package-comments
package pipeline

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/katastroma/keleustes"

	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/render"
	"github.com/katastroma/phortizo/internal/source"
)

// HandleMatch processes a matched webhook event through the pipeline.
func (r *Runner) HandleMatch(ctx context.Context, namespace string, m match.Result) {
	ctx, span := tracer.Start(ctx, "pipeline.run", trace.WithAttributes(
		attribute.String("tenant", namespace),
		attribute.String("watch_target.repo_url", m.Target.RepoURL),
		attribute.String("watch_target.ref", m.Target.Ref),
		attribute.String("watch_target.path", m.Target.Path),
	))
	defer span.End()

	// TODO Retrieve credentials from k8s
	// NOTE Use annotation from returned tenant watch target configmap

	// TODO Parse credentials for type

	// If necessary, exchange credentials for token

	var authMethod transport.AuthMethod

	fs, err := git.Clone(ctx, m.Target.RepoURL, m.Target.Ref, authMethod)
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

	connOpts := grpclib.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpclib.NewClient(rendererAddr, connOpts)
	if err != nil {
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"connecting to renderer failed",
			"tenant", namespace,
			"renderer", string(rendererType),
			"address", rendererAddr,
			"error", err,
		)
		return
	}
	defer conn.Close()

	client := pb.NewRendererServiceClient(conn)
	if err = render.Stream(ctx, client, fs, m.Target.Path); err != nil {
		span.RecordError(err)
		r.log.ErrorContext(
			ctx,
			"streaming to renderer failed",
			"tenant", namespace,
			"renderer", string(rendererType),
			"address", rendererAddr,
			"error", err,
		)
		return
	}

	r.log.InfoContext(ctx, "source streamed to renderer",
		"tenant", namespace,
		"renderer", string(rendererType),
	)
}
