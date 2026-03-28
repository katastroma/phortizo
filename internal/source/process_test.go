package source_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	mocktracer "git.sonicoriginal.software/grpc-testing/mocks/tracer"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"

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

func testTarget(credentialSecret string) *source.Target {
	return &source.Target{
		Name:             "wt-1",
		RepoURL:          "https://github.com/acme/app.git",
		Ref:              "refs/heads/main",
		Path:             "deploy/",
		CredentialSecret: credentialSecret,
	}
}

func succeedingAcquire(_ context.Context, _, _, _ string, _ int) error {
	return nil
}

func succeedingResolve(_ context.Context, _, _ string) (transport.AuthMethod, error) {
	return nil, nil
}

func succeedingVerify(_ context.Context, _, _, _ string) (bool, error) {
	return true, nil
}

func succeedingStream(_ context.Context, _ billy.Filesystem, _, _ string) error {
	return nil
}

func TestProcess_PublicRepo(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	lookupFn := func(_ billy.Filesystem, _ string) (string, error) {
		return "helm-renderer:8080", nil
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, lookupFn,
		succeedingVerify, succeedingStream,
	)
}

func TestProcess_PrivateRepo(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	lookupFn := func(_ billy.Filesystem, _ string) (string, error) {
		return "helm-renderer:8080", nil
	}

	testTarget("my-cred").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, lookupFn,
		succeedingVerify, succeedingStream,
	)
}

func TestProcess_LeaseAcquisitionError(t *testing.T) {
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	failingAcquire := func(_ context.Context, _, _, _ string, _ int) error {
		return fmt.Errorf("configmap not found")
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		failingAcquire, succeedingResolve, nil, nil, nil, nil,
	)
}

func TestProcess_CredentialRetrievalError(t *testing.T) {
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	failingResolve := func(_ context.Context, _, _ string) (transport.AuthMethod, error) {
		return nil, fmt.Errorf("not found")
	}

	testTarget("my-cred").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, failingResolve, nil, nil, nil, nil,
	)
}

func TestProcess_CloneError(t *testing.T) {
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	failingClone := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return nil, fmt.Errorf("clone failed")
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, failingClone, nil, nil, nil,
	)
}

func TestProcess_RendererNotConfigured(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	failingLookup := func(_ billy.Filesystem, _ string) (string, error) {
		return "", fmt.Errorf("no renderer configured for type %q", "kustomize")
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, failingLookup, nil, nil,
	)
}

func TestProcess_LeaseCheckError(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	lookupFn := func(_ billy.Filesystem, _ string) (string, error) {
		return "helm-renderer:8080", nil
	}
	failingVerify := func(_ context.Context, _, _, _ string) (bool, error) {
		return false, fmt.Errorf("configmap not found")
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, lookupFn,
		failingVerify, succeedingStream,
	)
}

func TestProcess_LeaseLost(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	lookupFn := func(_ billy.Filesystem, _ string) (string, error) {
		return "helm-renderer:8080", nil
	}
	lostVerify := func(_ context.Context, _, _, _ string) (bool, error) {
		return false, nil
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, lookupFn,
		lostVerify, succeedingStream,
	)
}

func TestProcess_StreamError(t *testing.T) {
	fs := helmFS(t)
	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)

	cloneFn := func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) {
		return fs, nil
	}
	lookupFn := func(_ billy.Filesystem, _ string) (string, error) {
		return "helm-renderer:8080", nil
	}
	failingStream := func(_ context.Context, _ billy.Filesystem, _, _ string) error {
		return fmt.Errorf("render failed")
	}

	testTarget("").Process(
		ctx, slog.Default(), mock.Tracer("test"), "tenant-a", 0,
		succeedingAcquire, succeedingResolve, cloneFn, lookupFn,
		succeedingVerify, failingStream,
	)
}
