//revive:disable:package-comments
package push

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/trace/noop"

	keleustes "github.com/katastroma/keleustes"

	"github.com/katastroma/phortizo/internal/source"
)

func helmFS(t *testing.T) billy.Filesystem {
	t.Helper()

	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}

	return fs
}

func noopClosures(t *testing.T) (
	func(context.Context, string, string, string, int) error,
	func(context.Context, string, string) (transport.AuthMethod, error),
	func(context.Context, string, string, transport.AuthMethod) (billy.Filesystem, error),
	func(context.Context, string, string, string) (bool, error),
	func(context.Context, billy.Filesystem, string, keleustes.RendererType) error,
) {
	t.Helper()

	fs := helmFS(t)

	return func(_ context.Context, _, _, _ string, _ int) error { return nil },
		func(_ context.Context, _, _ string) (transport.AuthMethod, error) { return nil, nil },
		func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) { return fs, nil },
		func(_ context.Context, _, _, _ string) (bool, error) { return true, nil },
		func(_ context.Context, _ billy.Filesystem, _ string, _ keleustes.RendererType) error { return nil }
}

func testHandler(t *testing.T, overrides ...func(*Handler)) *Handler {
	t.Helper()

	acquire, resolve, clone, verify, stream := noopClosures(t)
	h := &Handler{
		log:          slog.Default(),
		sem:          make(chan struct{}, 10),
		acquireLease: acquire,
		resolveAuth:  resolve,
		clone:        clone,
		verifyLease:  verify,
		stream:       stream,
	}

	for _, override := range overrides {
		override(h)
	}

	return h
}

func testTargets() []*source.Target {
	return []*source.Target{
		{
			Name:    "wt-1",
			RepoURL: "https://github.com/acme/app.git",
			Ref:     "refs/heads/main",
			Path:    "deploy/",
		},
	}
}

func TestProcessTargets(t *testing.T) {
	var processed int

	h := testHandler(t, func(h *Handler) {
		h.stream = func(_ context.Context, _ billy.Filesystem, _ string, _ keleustes.RendererType) error {
			processed++
			return nil
		}
	})

	tp := noop.NewTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(t.Context(), "test-event")

	h.processTargets(t.Context(), slog.Default(), tracer, "tenant-a", testTargets(), span)

	if processed != 1 {
		t.Errorf("expected 1 Process call, got %d", processed)
	}
}

func TestProcessTargets_MultipleTargets(t *testing.T) {
	var processed int

	h := testHandler(t, func(h *Handler) {
		h.stream = func(_ context.Context, _ billy.Filesystem, _ string, _ keleustes.RendererType) error {
			processed++
			return nil
		}
	})

	targets := []*source.Target{
		{Name: "wt-1", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		{Name: "wt-2", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "infra/"},
	}

	tp := noop.NewTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(t.Context(), "test-event")

	h.processTargets(t.Context(), slog.Default(), tracer, "tenant-a", targets, span)

	if processed != 2 {
		t.Errorf("expected 2 Process calls, got %d", processed)
	}
}

func TestProcessTargets_ErrorDoesNotBlockOthers(t *testing.T) {
	var processed int

	h := testHandler(t, func(h *Handler) {
		h.acquireLease = func(_ context.Context, _, name, _ string, _ int) error {
			if name == "wt-1" {
				return fmt.Errorf("lease failed")
			}
			return nil
		}
		h.stream = func(_ context.Context, _ billy.Filesystem, _ string, _ keleustes.RendererType) error {
			processed++
			return nil
		}
	})

	targets := []*source.Target{
		{Name: "wt-1", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		{Name: "wt-2", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "infra/"},
	}

	tp := noop.NewTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(t.Context(), "test-event")

	h.processTargets(t.Context(), slog.Default(), tracer, "tenant-a", targets, span)

	if processed != 1 {
		t.Errorf("expected 1 Process call (wt-2 only), got %d", processed)
	}
}

func TestProcessTargets_Semaphore(t *testing.T) {
	var maxConcurrent int
	var current int
	var mu sync.Mutex

	h := testHandler(t, func(h *Handler) {
		h.sem = make(chan struct{}, 1)
		h.stream = func(_ context.Context, _ billy.Filesystem, _ string, _ keleustes.RendererType) error {
			mu.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			runtime.Gosched()

			mu.Lock()
			current--
			mu.Unlock()
			return nil
		}
	})

	targets := []*source.Target{
		{Name: "wt-1", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		{Name: "wt-2", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
		{Name: "wt-3", RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}

	tp := noop.NewTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(t.Context(), "test-event")

	h.processTargets(t.Context(), slog.Default(), tracer, "tenant-a", targets, span)

	if maxConcurrent > 1 {
		t.Errorf("expected max 1 concurrent, got %d", maxConcurrent)
	}
}
