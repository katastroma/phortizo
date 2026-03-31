//revive:disable:package-comments
package retriever

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
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

func retrieveSourceTargetCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "st-1",
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
	k8s := fake.NewSimpleClientset(retrieveSourceTargetCM())

	var processedNamespace string
	var processedReplayCount int
	var processedTarget string

	acquireFn := func(_ context.Context, ns, name, _ string, rc int) error {
		processedNamespace = ns
		processedReplayCount = rc
		processedTarget = name
		return nil
	}
	resolveFn := func(_ context.Context, _, _ string) (transport.AuthMethod, error) { return nil, nil }
	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return helmFS(t), nil
	}
	verifyFn := func(_ context.Context, _, _, _ string) (bool, error) { return true, nil }
	streamFn := func(_ context.Context, _ billy.Filesystem, _ string) error { return nil }

	handler := New(
		slog.Default(), nil, k8s,
		10*time.Minute, 3,
		acquireFn, resolveFn, cloneFn, verifyFn, streamFn,
	)

	resp, err := handler.Retrieve(t.Context(), &pb.RetrieveRequest{
		Namespace:      "tenant-a",
		SourceTargetId: "st-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if processedNamespace != "tenant-a" {
		t.Errorf("namespace = %q, want %q", processedNamespace, "tenant-a")
	}

	if processedTarget != "st-1" {
		t.Errorf("target = %q, want %q", processedTarget, "st-1")
	}

	if processedReplayCount != 0 {
		t.Errorf("replayCount = %d, want %d", processedReplayCount, 0)
	}
}

