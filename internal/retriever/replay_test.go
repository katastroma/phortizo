//revive:disable:package-comments
package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/katastroma/naukleros"

	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tracing"
)

func helmFS(t *testing.T) billy.Filesystem {
	t.Helper()

	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}

	return fs
}

type processRecorder struct {
	calls       int
	namespace   string
	replayCount int
}

func (r *processRecorder) noopClosures(t *testing.T) (
	func(context.Context, string, string, string, int) error,
	func(context.Context, string, string) (transport.AuthMethod, error),
	func(context.Context, string, string, transport.AuthMethod) (billy.Filesystem, error),
	func(context.Context, string, string, string) (bool, error),
	func(context.Context, billy.Filesystem, string) error,
) {
	t.Helper()

	fs := helmFS(t)

	return func(_ context.Context, ns, _, _ string, rc int) error {
			r.namespace = ns
			r.replayCount = rc
			return nil
		},
		func(_ context.Context, _, _ string) (transport.AuthMethod, error) { return nil, nil },
		func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) { return fs, nil },
		func(_ context.Context, _, _, _ string) (bool, error) { return true, nil },
		func(_ context.Context, _ billy.Filesystem, _ string) error {
			r.calls++
			return nil
		}
}

func replayTrace() tracing.Trace {
	return tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.SourceTargetSpanName: {
			{
				tracing.SourceTargetNameAttribute:    "wt-1",
				tracing.SourceTargetRepoURLAttribute: "https://github.com/acme/app.git",
				tracing.SourceTargetRefAttribute:     "refs/heads/main",
				tracing.SourceTargetPathAttribute:    "deploy/",
			},
		},
	}
}

func sourceTargetCM() *corev1.ConfigMap {
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
	k8s := fake.NewSimpleClientset(sourceTargetCM())
	rec := &processRecorder{}
	acquire, resolve, clone, verify, stream := rec.noopClosures(t)

	handler := New(
		slog.Default(), &stubTracer{trace: replayTrace()}, k8s,
		10*time.Minute, 3,
		acquire, resolve, clone, verify, stream,
	)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if rec.calls != 1 {
		t.Fatalf("expected 1 Process call, got %d", rec.calls)
	}

	if rec.namespace != "tenant-a" {
		t.Errorf("namespace = %q, want %q", rec.namespace, "tenant-a")
	}

	if rec.replayCount != 1 {
		t.Errorf("replayCount = %d, want %d", rec.replayCount, 1)
	}
}

func TestReplay_ActiveLease(t *testing.T) {
	cm := sourceTargetCM()
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "other-lease",
		lease.StartedAnnotation:     time.Now().UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	rec := &processRecorder{}
	acquire, resolve, clone, verify, stream := rec.noopClosures(t)

	handler := New(
		slog.Default(), &stubTracer{trace: replayTrace()}, k8s,
		10*time.Minute, 3,
		acquire, resolve, clone, verify, stream,
	)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if rec.calls != 0 {
		t.Errorf("expected no Process calls, got %d", rec.calls)
	}
}

func TestReplay_StaleLease(t *testing.T) {
	cm := sourceTargetCM()
	staleTime := time.Now().Add(-20 * time.Minute)
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "old-lease",
		lease.StartedAnnotation:     staleTime.UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "0",
	}
	k8s := fake.NewSimpleClientset(cm)
	rec := &processRecorder{}
	acquire, resolve, clone, verify, stream := rec.noopClosures(t)

	handler := New(
		slog.Default(), &stubTracer{trace: replayTrace()}, k8s,
		10*time.Minute, 3,
		acquire, resolve, clone, verify, stream,
	)

	resp, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if rec.calls != 1 {
		t.Fatalf("expected 1 Process call, got %d", rec.calls)
	}
}

func TestReplay_MaxReplayAttempts(t *testing.T) {
	cm := sourceTargetCM()
	staleTime := time.Now().Add(-20 * time.Minute)
	cm.Annotations = map[string]string{
		lease.IDAnnotation:          "old-lease",
		lease.StartedAnnotation:     staleTime.UTC().Format(time.RFC3339),
		lease.ReplayCountAnnotation: "3",
	}
	k8s := fake.NewSimpleClientset(cm)

	handler := New(
		slog.Default(), &stubTracer{trace: replayTrace()}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	_, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"})
	if err == nil {
		t.Fatal("expected error for max replay attempts")
	}
}

func TestReplay_TraceQueryError(t *testing.T) {
	st := &stubTracer{err: fmt.Errorf("tempo unavailable")}
	handler := New(
		slog.Default(), st, nil,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error when trace query fails")
	}
}

func TestReplay_NoEventSpan(t *testing.T) {
	tr := tracing.Trace{tracing.SourceTargetSpanName: {{tracing.SourceTargetNameAttribute: "wt-1"}}}
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, nil,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing event span")
	}
}

func TestReplay_MissingTenant(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:        {{}},
		tracing.SourceTargetSpanName: {{tracing.SourceTargetNameAttribute: "wt-1"}},
	}
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, nil,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing tenant attribute")
	}
}

func TestReplay_MissingName(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:        {{tracing.TenantAttribute: "tenant-a"}},
		tracing.SourceTargetSpanName: {{}},
	}
	k8s := fake.NewSimpleClientset(sourceTargetCM())
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing name attribute")
	}
}

func TestReplay_MissingRepoURL(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName:        {{tracing.TenantAttribute: "tenant-a"}},
		tracing.SourceTargetSpanName: {{tracing.SourceTargetNameAttribute: "wt-1"}},
	}
	k8s := fake.NewSimpleClientset(sourceTargetCM())
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing repo_url attribute")
	}
}

func TestReplay_MissingRef(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.SourceTargetSpanName: {
			{
				tracing.SourceTargetNameAttribute:    "wt-1",
				tracing.SourceTargetRepoURLAttribute: "https://github.com/acme/app.git",
			},
		},
	}
	k8s := fake.NewSimpleClientset(sourceTargetCM())
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing ref attribute")
	}
}

func TestReplay_MissingPath(t *testing.T) {
	tr := tracing.Trace{
		tracing.EventSpanName: {{tracing.TenantAttribute: "tenant-a"}},
		tracing.SourceTargetSpanName: {
			{
				tracing.SourceTargetNameAttribute:    "wt-1",
				tracing.SourceTargetRepoURLAttribute: "https://github.com/acme/app.git",
				tracing.SourceTargetRefAttribute:     "refs/heads/main",
			},
		},
	}
	k8s := fake.NewSimpleClientset(sourceTargetCM())
	handler := New(
		slog.Default(), &stubTracer{trace: tr}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error for missing path attribute")
	}
}

func TestReplay_LeaseReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	handler := New(
		slog.Default(), &stubTracer{trace: replayTrace()}, k8s,
		10*time.Minute, 3,
		nil, nil, nil, nil, nil,
	)

	if _, err := handler.Replay(t.Context(), &pb.ReplayRequest{EventId: "trace-123"}); err == nil {
		t.Fatal("expected error when configmap not found")
	}
}
