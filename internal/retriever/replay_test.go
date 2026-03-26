//revive:disable:package-comments
package retriever

import (
	"fmt"
	"log/slog"
	"testing"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/tracequery"
)

func TestReplay(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"namespace":             "reg-1",
		"tenant":                "acme",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	handler := New(slog.Default(), querier, nil)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	querier := &stubQuerier{err: fmt.Errorf("tempo unavailable")}
	handler := New(slog.Default(), querier, nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_MissingAttributes(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"watch_target.repo_url": "https://github.com/acme/app.git",
	}}
	handler := New(slog.Default(), querier, nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing attributes")
	}
}

func TestReplay_MissingWatchTarget(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"namespace": "reg-1",
		"tenant":    "acme",
	}}
	handler := New(slog.Default(), querier, nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing watch target attributes")
	}
}

func TestReplay_RegistrationNotFound(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"namespace":             "nonexistent",
		"tenant":                "acme",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	handler := New(slog.Default(), querier, nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing registration")
	}
}
