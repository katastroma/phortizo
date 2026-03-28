//revive:disable:package-comments
package retriever

import (
	"context"
	"log/slog"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/katastroma/naukleros"

	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
)

type stubTracer struct {
	trace tracing.Trace
	err   error
}

func (q *stubTracer) GetTrace(_ context.Context, _ string) (tracing.Trace, error) {
	return q.trace, q.err
}

func retrieveWatchTargetCM() *corev1.ConfigMap {
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

func TestRetrieve(t *testing.T) {
	k8s := fake.NewSimpleClientset(retrieveWatchTargetCM())
	runner := &mockHandler{}
	handler := New(slog.Default(), nil, runner, k8s, 10*time.Minute, 3)

	resp, err := handler.Retrieve(t.Context(), &pb.RetrieveRequest{
		Namespace:     "tenant-a",
		WatchTargetId: "wt-1",
	})
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

	if runner.calls[0].Name != "wt-1" {
		t.Errorf("Name = %q, want %q", runner.calls[0].Name, "wt-1")
	}

	if runner.replayCount != 0 {
		t.Errorf("replayCount = %d, want %d", runner.replayCount, 0)
	}
}

func TestRetrieve_WatchTargetNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	handler := New(slog.Default(), nil, nil, k8s, 10*time.Minute, 3)

	_, err := handler.Retrieve(t.Context(), &pb.RetrieveRequest{
		Namespace:     "tenant-a",
		WatchTargetId: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for missing watch target")
	}
}
