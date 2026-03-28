//revive:disable:package-comments
package renderer_test

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"

	"github.com/katastroma/phortizo/internal/renderer"
)

func TestLookupFunc_Helm(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/Chart.yaml")

	renderers := map[renderer.Type]string{renderer.Helm: "helm-renderer:8080"}
	lookup := renderer.LookupFunc(renderers)

	addr, err := lookup(fs, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addr != "helm-renderer:8080" {
		t.Errorf("addr = %q, want %q", addr, "helm-renderer:8080")
	}
}

func TestLookupFunc_Kustomize(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/kustomization.yaml")

	renderers := map[renderer.Type]string{renderer.Kustomize: "kustomize-renderer:8080"}
	lookup := renderer.LookupFunc(renderers)

	addr, err := lookup(fs, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addr != "kustomize-renderer:8080" {
		t.Errorf("addr = %q, want %q", addr, "kustomize-renderer:8080")
	}
}

func TestLookupFunc_Raw(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/deployment.yaml")

	renderers := map[renderer.Type]string{renderer.Raw: "raw-renderer:8080"}
	lookup := renderer.LookupFunc(renderers)

	addr, err := lookup(fs, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if addr != "raw-renderer:8080" {
		t.Errorf("addr = %q, want %q", addr, "raw-renderer:8080")
	}
}

func TestLookupFunc_NotConfigured(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/Chart.yaml")

	renderers := map[renderer.Type]string{}
	lookup := renderer.LookupFunc(renderers)

	if _, err := lookup(fs, "deploy"); err == nil {
		t.Fatal("expected error for unconfigured renderer")
	}
}
