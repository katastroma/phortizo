//revive:disable:package-comments
package pipeline

import (
	"log/slog"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/renderer"
)

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, and stream to the renderer.
type Runner struct {
	log         *slog.Logger
	httpClient  *http.Client
	renderers   map[renderer.Type]string
	credentials credential.Reader
	cloner      git.Cloner
	renderer    renderer.Streamer
	k8sClient   kubernetes.Interface
}

// New creates a pipeline runner.
func New(
	log *slog.Logger,
	httpClient *http.Client,
	renderers map[renderer.Type]string,
	credentials credential.Reader,
	cloner git.Cloner,
	renderer renderer.Streamer,
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
