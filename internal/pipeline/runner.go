//revive:disable:package-comments
package pipeline

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/auth"
	"github.com/katastroma/phortizo/internal/source"
)

var tracer = otel.Tracer("pipeline")

// SpanName is the OTel span name used by the pipeline run. The retriever
// uses this to query trace attributes for replay.
const SpanName = "pipeline.run"

// CredentialReader reads credentials from a tenant namespace.
type CredentialReader interface {
	Get(ctx context.Context, namespace, name string) (auth.Credential, error)
}

// Cloner performs a git clone and returns the worktree filesystem.
type Cloner interface {
	Clone(ctx context.Context, url, ref string, auth transport.AuthMethod) (billy.Filesystem, error)
}

// Renderer streams source content to a renderer service.
type Renderer interface {
	Render(ctx context.Context, fs billy.Filesystem, path, addr string) error
}

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log         *slog.Logger
	httpClient  *http.Client
	renderers   map[source.RendererType]string
	credentials CredentialReader
	cloner      Cloner
	renderer    Renderer
	k8sClient   kubernetes.Interface
}

// New creates a pipeline runner.
func New(
	log *slog.Logger,
	httpClient *http.Client,
	renderers map[source.RendererType]string,
	credentials CredentialReader,
	cloner Cloner,
	renderer Renderer,
	k8sClient kubernetes.Interface,
) *Runner {
	return &Runner{
		log:         log,
		httpClient:  httpClient,
		renderers:   renderers,
		credentials: credentials,
		cloner:      cloner,
		renderer:    renderer,
		k8sClient:   k8sClient,
	}
}
