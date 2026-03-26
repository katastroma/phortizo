//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/tracequery"
)

type mockHandler struct {
	calls     []registration.WatchTarget
	namespace string
}

func (m *mockHandler) HandleMatch(_ context.Context, namespace string, r registration.WatchTarget) {
	m.namespace = namespace
	m.calls = append(m.calls, r)
}

func TestReplay(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"tenant":                "tenant-a",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	runner := &mockHandler{}
	handler := New(slog.Default(), querier, runner)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 HandleMatch call, got %d", len(runner.calls))
	}

	if runner.namespace != "tenant-a" {
		t.Errorf("namespace = %q, want %q", runner.namespace, "tenant-a")
	}

	if runner.calls[0].RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", runner.calls[0].RepoURL, "https://github.com/acme/app.git")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	querier := &stubQuerier{err: fmt.Errorf("tempo unavailable")}
	handler := New(slog.Default(), querier, nil)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_MissingTenant(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	handler := New(slog.Default(), querier, nil)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing tenant attribute")
	}
}

func TestReplay_MissingRepoURL(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"tenant":            "tenant-a",
		"watch_target.ref":  "refs/heads/main",
		"watch_target.path": "deploy/",
	}}
	handler := New(slog.Default(), querier, nil)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing repo_url attribute")
	}
}

func TestReplay_MissingRef(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"tenant":                "tenant-a",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.path":     "deploy/",
	}}
	handler := New(slog.Default(), querier, nil)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing ref attribute")
	}
}

func TestReplay_MissingPath(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"tenant":                "tenant-a",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
	}}
	handler := New(slog.Default(), querier, nil)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing path attribute")
	}
}
