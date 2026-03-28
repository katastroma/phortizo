//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/katastroma/naukleros"

	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
)

type mockHandler struct {
	calls       []*source.Target
	namespace   string
	replayCount int
}

func (m *mockHandler) HandleMatch(
	_ context.Context,
	_ trace.Tracer,
	namespace string,
	target *source.Target,
	replayCount int,
) {
	m.namespace = namespace
	m.replayCount = replayCount
	m.calls = append(m.calls, target)
}

func replayTrace() tracing.Trace {
	return tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.WatchTargetSpanName: {
			{
				tracing.WatchTargetNameAttribute:    "wt-1",
				tracing.WatchTargetRepoURLAttribute: "https://github.com/acme/app.git",
				tracing.WatchTargetRefAttribute:     "refs/heads/main",
				tracing.WatchTargetPathAttribute:    "deploy/",
			},
		},
	}
}

func watchTargetCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wt-1",
			Namespace: "tenant-a",
			Labels:    map[string]string{object.TypeLabel: source.TypeLabel},
		},
		Data: map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		},
	}
}

func TestReplay(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := &mockHandler{}
	handler := New(slog.Default(), &stubTracer{trace: replayTrace()}, runner, k8s, 10*time.Minute, 3)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
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
		t.Errorf(
			"namespace = %q, want %q",
			runner.namespace, "tenant-a",
		)
	}

	want := "https://github.com/acme/app.git"
	if runner.calls[0].RepoURL != want {
		t.Errorf("RepoURL = %q, want %q", runner.calls[0].RepoURL, want)
	}

	if runner.replayCount != 1 {
		t.Errorf("replayCount = %d, want %d", runner.replayCount, 1)
	}
}

func TestReplay_ActiveLease(t *testing.T) {
	cm := watchTargetCM()
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "other-lease",
		lease.StartedAnnotation:     time.Now().UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	runner := &mockHandler{}
	handler := New(slog.Default(), &stubTracer{trace: replayTrace()}, runner, k8s, 10*time.Minute, 3)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if len(runner.calls) != 0 {
		t.Errorf("expected no HandleMatch calls, got %d", len(runner.calls))
	}
}

func TestReplay_StaleLease(t *testing.T) {
	cm := watchTargetCM()
	staleTime := time.Now().Add(-20 * time.Minute)
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "old-lease",
		lease.StartedAnnotation:     staleTime.UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	runner := &mockHandler{}
	handler := New(slog.Default(), &stubTracer{trace: replayTrace()}, runner, k8s, 10*time.Minute, 3)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 HandleMatch call, got %d", len(runner.calls))
	}
}

func TestReplay_MaxReplayAttempts(t *testing.T) {
	cm := watchTargetCM()
	staleTime := time.Now().Add(-20 * time.Minute)
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "old-lease",
		lease.StartedAnnotation:     staleTime.UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "3",
	}
	k8s := fake.NewSimpleClientset(cm)
	handler := New(slog.Default(), &stubTracer{trace: replayTrace()}, nil, k8s, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for max replay attempts")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	st := &stubTracer{err: fmt.Errorf("tempo unavailable")}
	handler := New(slog.Default(), st, nil, nil, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_NoEventSpan(t *testing.T) {
	tr := tracing.Trace{tracing.WatchTargetSpanName: {{tracing.WatchTargetNameAttribute: "wt-1"}}}
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, nil, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing event span")
	}
}

func TestReplay_MissingTenant(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:       {{}},
		tracing.WatchTargetSpanName: {{tracing.WatchTargetNameAttribute: "wt-1"}},
	}
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, nil, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing tenant attribute")
	}
}

func TestReplay_MissingName(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:       {{tracing.TenantAttribute: "tenant-a"}},
		tracing.WatchTargetSpanName: {{}},
	}
	k8s := fake.NewSimpleClientset(watchTargetCM())
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, k8s, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing name attribute")
	}
}

func TestReplay_MissingRepoURL(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:       {{tracing.TenantAttribute: "tenant-a"}},
		tracing.WatchTargetSpanName: {{tracing.WatchTargetNameAttribute: "wt-1"}},
	}
	k8s := fake.NewSimpleClientset(watchTargetCM())
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, k8s, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing repo_url attribute")
	}
}

func TestReplay_MissingRef(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.WatchTargetSpanName: {
			{
				tracing.WatchTargetNameAttribute:    "wt-1",
				tracing.WatchTargetRepoURLAttribute: "https://github.com/acme/app.git",
			},
		},
	}
	k8s := fake.NewSimpleClientset(watchTargetCM())
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, k8s, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing ref attribute")
	}
}

func TestReplay_MissingPath(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.WatchTargetSpanName: {
			{
				tracing.WatchTargetNameAttribute:    "wt-1",
				tracing.WatchTargetRepoURLAttribute: "https://github.com/acme/app.git",
				tracing.WatchTargetRefAttribute:     "refs/heads/main",
			},
		},
	}
	k8s := fake.NewSimpleClientset(watchTargetCM())
	handler := New(slog.Default(), &stubTracer{trace: tr}, nil, k8s, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing path attribute")
	}
}

func TestReplay_LeaseReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	handler := New(slog.Default(), &stubTracer{trace: replayTrace()}, nil, k8s, 10*time.Minute, 3)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error when configmap not found")
	}
}
