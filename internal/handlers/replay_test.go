package handler

import (
	"fmt"
	"log/slog"
	"testing"

	pb "github.com/katastroma/naukleros"
	regmemory "github.com/katastroma/phortizo/internal/registration/memory"
	"github.com/katastroma/phortizo/internal/tracequery"
)

func TestReplay(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"registration_id":       "reg-1",
		"tenant":                "acme",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	store := regmemory.New()
	store.Add(testRegistration())
	capture := &captureMatch{}
	handler := NewRetriever(slog.Default(), querier, store, capture.handle)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	capture.mu.Lock()
	defer capture.mu.Unlock()
	if len(capture.results) != 1 {
		t.Fatalf("expected 1 match dispatched, got %d", len(capture.results))
	}
	if capture.results[0].Registration.TenantID != "acme" {
		t.Errorf("tenant = %q, want %q", capture.results[0].Registration.TenantID, "acme")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	querier := &stubQuerier{err: fmt.Errorf("tempo unavailable")}
	handler := NewRetriever(slog.Default(), querier, regmemory.New(), nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_MissingAttributes(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"watch_target.repo_url": "https://github.com/acme/app.git",
	}}
	handler := NewRetriever(slog.Default(), querier, regmemory.New(), nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing attributes")
	}
}

func TestReplay_MissingWatchTarget(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"registration_id": "reg-1",
		"tenant":          "acme",
	}}
	handler := NewRetriever(slog.Default(), querier, regmemory.New(), nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing watch target attributes")
	}
}

func TestReplay_RegistrationNotFound(t *testing.T) {
	querier := &stubQuerier{attrs: tracequery.Attributes{
		"registration_id":       "nonexistent",
		"tenant":                "acme",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}}
	handler := NewRetriever(slog.Default(), querier, regmemory.New(), nil)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing registration")
	}
}
