//revive:disable:package-comments
package pipeline

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/katastroma/keleusma/broker"
	"github.com/katastroma/kytos/store"

	"github.com/katastroma/phortizo/internal/archive"
	"github.com/katastroma/phortizo/internal/auth"
	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/source"
)

// Runner orchestrates the pipeline for a matched webhook event: resolve
// credentials, clone, inspect, archive, store, and publish.
type Runner struct {
	log         *slog.Logger
	credentials credential.Store
	exchanger   auth.TokenExchanger
	storage     store.Store
	publisher   broker.Publisher
}

// NewRunner creates a pipeline runner.
func NewRunner(
	log *slog.Logger,
	credentials credential.Store,
	exchanger auth.TokenExchanger,
	storage store.Store,
	publisher broker.Publisher,
) *Runner {
	return &Runner{
		log:         log,
		credentials: credentials,
		exchanger:   exchanger,
		storage:     storage,
		publisher:   publisher,
	}
}

// HandleMatch processes a matched webhook event through the full pipeline.
func (r *Runner) HandleMatch(mc match.Result) {
	runID, err := generateRunID()
	if err != nil {
		r.log.Error("generating run id", "error", err)
		return
	}

	tenant := mc.Registration.TenantID
	ctx := context.Background()

	cred, err := r.credentials.Get(ctx, mc.Registration.CredentialRef)
	if err != nil {
		r.log.Error("resolving credentials", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	authMethod, err := auth.Resolve(ctx, cred, r.exchanger)
	if err != nil {
		r.log.Error("resolving auth", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	fs, err := git.Clone(ctx, mc.Target.RepoURL, mc.Target.Ref, authMethod)
	if err != nil {
		r.log.Error("cloning repo", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	rendererType := source.DetectRenderer(fs, mc.Target.Path)

	archiveReader, err := archive.FromBillyFS(fs, mc.Target.Path)
	if err != nil {
		r.log.Error("archiving source", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	storageKey := fmt.Sprintf("pipeline/%s/%s/source", tenant, runID)
	if err := r.storage.Put(ctx, storageKey, archiveReader); err != nil {
		r.log.Error("writing to storage", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	r.publish(ctx, tenant, runID, storageKey, rendererType)
}

func (r *Runner) publish(ctx context.Context, tenant, runID, storageKey string, rendererType source.RendererType) {
	event, err := json.Marshal(sourceReadyEvent{
		RunID:        runID,
		Tenant:       tenant,
		StorageKey:   storageKey,
		RendererType: string(rendererType),
	})
	if err != nil {
		r.log.Error("marshaling event", "tenant", tenant, "run_id", runID, "error", err)
		return
	}

	subject := fmt.Sprintf("source.%s", tenant)
	if err := r.publisher.Publish(ctx, subject, event); err != nil {
		r.log.Error("publishing source ready event", "tenant", tenant, "run_id", runID, "error", err)
	}
}

type sourceReadyEvent struct {
	RunID        string `json:"run_id"`
	Tenant       string `json:"tenant"`
	StorageKey   string `json:"storage_key"`
	RendererType string `json:"renderer_type"`
}

func generateRunID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
