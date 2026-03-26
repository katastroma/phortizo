package pipeline_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/katastroma/phortizo/internal/auth"
	"github.com/katastroma/phortizo/internal/match"
	"github.com/katastroma/phortizo/internal/pipeline"
	"github.com/katastroma/phortizo/internal/registration"
	"github.com/katastroma/phortizo/internal/source"
)

type mockCredentialReader struct {
	cred auth.Credential
	err  error
}

func (m *mockCredentialReader) Get(context.Context, string, string) (auth.Credential, error) {
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

func (m *mockCloner) Clone(context.Context, string, string, transport.AuthMethod) (billy.Filesystem, error) {
	return m.fs, m.err
}

type mockRenderer struct {
	err error
}

func (m *mockRenderer) Render(context.Context, billy.Filesystem, string, string) error {
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

func testRunner(
	credentials pipeline.CredentialReader,
	cloner pipeline.Cloner,
	renderer pipeline.Renderer,
) *pipeline.Runner {
	renderers := map[source.RendererType]string{
		source.Helm: "helm-renderer:8080",
	}

	return pipeline.New(slog.Default(), renderers, credentials, http.DefaultClient, cloner, renderer)
}

func testResult(credentialSecret string) match.Result {
	return match.Result{
		Target: registration.WatchTarget{
			RepoURL:          "https://github.com/acme/app.git",
			Ref:              "refs/heads/main",
			Path:             "deploy/",
			CredentialSecret: credentialSecret,
		},
	}
}

func TestHandleMatch_PublicRepo(t *testing.T) {
	runner := testRunner(
		nil,
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult(""))
}

func TestHandleMatch_PrivateRepo(t *testing.T) {
	runner := testRunner(
		&mockCredentialReader{cred: &mockCredential{}},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult("my-cred"))
}

func TestHandleMatch_CredentialRetrievalError(t *testing.T) {
	runner := testRunner(
		&mockCredentialReader{err: fmt.Errorf("not found")},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult("my-cred"))
}

func TestHandleMatch_AuthenticateError(t *testing.T) {
	runner := testRunner(
		&mockCredentialReader{cred: &mockCredential{err: fmt.Errorf("auth failed")}},
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult("my-cred"))
}

func TestHandleMatch_CloneError(t *testing.T) {
	runner := testRunner(
		nil,
		&mockCloner{err: fmt.Errorf("clone failed")},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult(""))
}

func TestHandleMatch_RendererNotConfigured(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/kustomization.yaml", []byte("resources: []"), 0o644); err != nil {
		t.Fatalf("writing kustomization.yaml: %v", err)
	}

	runner := testRunner(
		nil,
		&mockCloner{fs: fs},
		&mockRenderer{},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult(""))
}

func TestHandleMatch_RenderError(t *testing.T) {
	runner := testRunner(
		nil,
		&mockCloner{fs: helmFS(t)},
		&mockRenderer{err: fmt.Errorf("render failed")},
	)

	runner.HandleMatch(t.Context(), "tenant-a", testResult(""))
}
