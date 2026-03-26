//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/katastroma/naukleros"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/tracequery"
)

type mockHandler struct {
	calls        []registration.WatchTarget
	namespace    string
	replayCount  int
}

func (m *mockHandler) HandleMatch(_ context.Context, namespace string, target registration.WatchTarget, replayCount int) {
	m.namespace = namespace
	m.replayCount = replayCount
	m.calls = append(m.calls, target)
}

func replayAttrs() tracequery.Attributes {
	return tracequery.Attributes{
		"tenant":                "tenant-a",
		"watch_target.name":    "wt-1",
		"watch_target.repo_url": "https://github.com/acme/app.git",
		"watch_target.ref":      "refs/heads/main",
		"watch_target.path":     "deploy/",
	}
}

func watchTargetCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wt-1",
			Namespace: "tenant-a",
			Labels:    map[string]string{configmap.TypeLabel: configmap.WatchTargetType},
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
	handler := New(slog.Default(), &stubQuerier{attrs: replayAttrs()}, runner, k8s, 10*time.Minute, 3)

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

	if runner.replayCount != 1 {
		t.Errorf("replayCount = %d, want %d", runner.replayCount, 1)
	}
}

func TestReplay_ActiveLease(t *testing.T) {
	cm := watchTargetCM()
	cm.Annotations = map[string]string{
		configmap.LeaseRunIDAnnotation:       "other-run",
		configmap.LeaseStartedAnnotation:     time.Now().UTC().Format(time.RFC3339),
		configmap.LeaseReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	runner := &mockHandler{}
	handler := New(slog.Default(), &stubQuerier{attrs: replayAttrs()}, runner, k8s, 10*time.Minute, 3)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
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
	cm.Annotations = map[string]string{
		configmap.LeaseRunIDAnnotation:       "old-run",
		configmap.LeaseStartedAnnotation:     time.Now().Add(-20 * time.Minute).UTC().Format(time.RFC3339),
		configmap.LeaseReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	runner := &mockHandler{}
	handler := New(slog.Default(), &stubQuerier{attrs: replayAttrs()}, runner, k8s, 10*time.Minute, 3)

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
}

func TestReplay_MaxReplayAttempts(t *testing.T) {
	cm := watchTargetCM()
	cm.Annotations = map[string]string{
		configmap.LeaseRunIDAnnotation:       "old-run",
		configmap.LeaseStartedAnnotation:     time.Now().Add(-20 * time.Minute).UTC().Format(time.RFC3339),
		configmap.LeaseReplayCountAnnotation: "3",
	}
	k8s := fake.NewSimpleClientset(cm)
	handler := New(slog.Default(), &stubQuerier{attrs: replayAttrs()}, nil, k8s, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for max replay attempts")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	handler := New(slog.Default(), &stubQuerier{err: fmt.Errorf("tempo unavailable")}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_MissingTenant(t *testing.T) {
	attrs := replayAttrs()
	delete(attrs, "tenant")
	handler := New(slog.Default(), &stubQuerier{attrs: attrs}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing tenant attribute")
	}
}

func TestReplay_MissingName(t *testing.T) {
	attrs := replayAttrs()
	delete(attrs, "watch_target.name")
	handler := New(slog.Default(), &stubQuerier{attrs: attrs}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing name attribute")
	}
}

func TestReplay_MissingRepoURL(t *testing.T) {
	attrs := replayAttrs()
	delete(attrs, "watch_target.repo_url")
	handler := New(slog.Default(), &stubQuerier{attrs: attrs}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing repo_url attribute")
	}
}

func TestReplay_MissingRef(t *testing.T) {
	attrs := replayAttrs()
	delete(attrs, "watch_target.ref")
	handler := New(slog.Default(), &stubQuerier{attrs: attrs}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing ref attribute")
	}
}

func TestReplay_MissingPath(t *testing.T) {
	attrs := replayAttrs()
	delete(attrs, "watch_target.path")
	handler := New(slog.Default(), &stubQuerier{attrs: attrs}, nil, nil, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for missing path attribute")
	}
}

func TestReplay_LeaseReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	handler := New(slog.Default(), &stubQuerier{attrs: replayAttrs()}, nil, k8s, 10*time.Minute, 3)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{RunId: "trace-123"})
	if err == nil {
		t.Fatal("expected error when configmap not found")
	}
}
