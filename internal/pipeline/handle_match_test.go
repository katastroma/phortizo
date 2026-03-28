package pipeline_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	mocktracer "git.sonicoriginal.software/grpc-testing/mocks/tracer"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/git"
	"github.com/katastroma/phortizo/internal/k8s"
	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/lease"
	"github.com/katastroma/phortizo/internal/pipeline"
	"github.com/katastroma/phortizo/internal/renderer"
	"github.com/katastroma/phortizo/internal/source"
)

type mockCredentialReader struct {
	cred credential.Authenticator
	err  error
}

func (m *mockCredentialReader) Get(context.Context, string, string) (credential.Authenticator, error) {
	return m.cred, m.err
}

type mockCredential struct {
	authMethod transport.AuthMethod
	err        error
}

func (m *mockCredential) Authenticate(context.Context, *http.Client) (transport.AuthMethod, error) {
	return m.authMethod, m.err
}

type mockCloner struct {
	fs  billy.Filesystem
	err error
}

func (m *mockCloner) Clone(
	context.Context, string, string, transport.AuthMethod,
) (billy.Filesystem, error) {
	return m.fs, m.err
}

type configMapDeletingCloner struct {
	fs        billy.Filesystem
	k8sClient *fake.Clientset
}

func (c *configMapDeletingCloner) Clone(
	ctx context.Context, _, _ string, _ transport.AuthMethod,
) (billy.Filesystem, error) {
	err := c.k8sClient.CoreV1().ConfigMaps("tenant-a").Delete(ctx, "wt-1", metav1.DeleteOptions{})
	if err != nil {
		return nil, fmt.Errorf("deleting configmap: %w", err)
	}

	return c.fs, nil
}

type mockRenderer struct {
	err error
}

func (m *mockRenderer) Stream(context.Context, billy.Filesystem, string, string) error {
	return m.err
}

func helmFS(t *testing.T) billy.Filesystem {
	t.Helper()
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}

	return fs
}

func watchTargetCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wt-1",
			Namespace: "tenant-a",
			Labels:    map[string]string{k8s.TypeLabel: source.TypeLabel},
		},
		Data: map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		},
	}
}

func testRunner(
	k8sClient *fake.Clientset,
	credentials credential.Reader,
	cloner git.Cloner,
	r renderer.Streamer,
) *pipeline.Runner {
	renderers := map[renderer.Type]string{renderer.Helm: "helm-renderer:8080"}
	return pipeline.New(
		slog.Default(),
		http.DefaultClient,
		renderers, credentials,
		cloner,
		r,
		k8sClient,
	)
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

func TestHandleMatch_PublicRepo(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(k8s, nil, &mockCloner{fs: helmFS(t)}, &mockRenderer{})

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

func TestHandleMatch_PrivateRepo(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(
		k8s,
		&mockCredentialReader{cred: &mockCredential{}},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget("my-cred"), 0)
}

func TestHandleMatch_CredentialRetrievalError(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(
		k8s,
		&mockCredentialReader{err: fmt.Errorf("not found")},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget("my-cred"), 0)
}

func TestHandleMatch_AuthenticateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(
		k8s,
		&mockCredentialReader{cred: &mockCredential{err: fmt.Errorf("auth failed")}},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget("my-cred"), 0)
}

func TestHandleMatch_CloneError(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(
		k8s,
		nil,
		&mockCloner{err: fmt.Errorf("clone failed")},
		&mockRenderer{},
	)

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

func TestHandleMatch_RendererNotConfigured(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	fs := memfs.New()
	if err := util.WriteFile(
		fs,
		"deploy/kustomization.yaml",
		[]byte("resources: []"),
		0o644,
	); err != nil {
		t.Fatalf("writing kustomization.yaml: %v", err)
	}

	runner := testRunner(k8s, nil, &mockCloner{fs: fs}, &mockRenderer{})

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

func TestHandleMatch_RenderError(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())
	runner := testRunner(
		k8s,
		nil,
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{err: fmt.Errorf("render failed")},
	)

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

func TestHandleMatch_LeaseAcquisitionError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	runner := testRunner(k8s, nil, &mockCloner{fs: helmFS(t)}, &mockRenderer{})

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

func TestHandleMatch_LeaseLost(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())

	// Use a cloner that clobbers the lease mid-pipeline (simulating another
	// run taking ownership between clone and render).
	clobberer := &leaseStealingCloner{
		fs:        helmFS(t),
		k8sClient: k8s,
	}
	runner := testRunner(k8s, nil, clobberer, &mockRenderer{})

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}

type leaseStealingCloner struct {
	fs        billy.Filesystem
	k8sClient *fake.Clientset
}

func (c *leaseStealingCloner) Clone(
	ctx context.Context,
	_, _ string,
	_ transport.AuthMethod,
) (billy.Filesystem, error) {
	// After the clone "succeeds", another run steals the lease
	store := configmap.NewStore(c.k8sClient, "tenant-a")
	err := lease.Acquire(ctx, store, "wt-1", "different-run", 0)
	if err != nil {
		return nil, fmt.Errorf("stealing lease: %w", err)
	}

	return c.fs, nil
}

func TestHandleMatch_LeaseCheckError(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetCM())

	// Cloner deletes the ConfigMap mid-pipeline so the lease check fails
	deleter := &configMapDeletingCloner{
		fs:        helmFS(t),
		k8sClient: k8s,
	}
	runner := testRunner(k8s, nil, deleter, &mockRenderer{})

	mock, ctx := mocktracer.New(t)
	defer mock.Shutdown(t)
	runner.HandleMatch(ctx, mock.Tracer("test"), "tenant-a", testTarget(""), 0)
}
