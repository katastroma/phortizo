package renderer_test

import (
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"

	"github.com/katastroma/keleustes"
	"github.com/katastroma/phortizo/internal/renderer"
)

func TestDetect_Helm(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_HELM {
		t.Errorf("expected HELM, got %v", got)
	}
}

func TestDetect_KustomizationYaml(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/kustomization.yaml", []byte("resources:"), 0o644); err != nil {
		t.Fatalf("writing kustomization.yaml: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_KUSTOMIZE {
		t.Errorf("expected KUSTOMIZE, got %v", got)
	}
}

func TestDetect_KustomizationYml(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/kustomization.yml", []byte("resources:"), 0o644); err != nil {
		t.Fatalf("writing kustomization.yml: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_KUSTOMIZE {
		t.Errorf("expected KUSTOMIZE, got %v", got)
	}
}

func TestDetect_KustomizationCapitalized(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Kustomization", []byte("resources:"), 0o644); err != nil {
		t.Fatalf("writing Kustomization: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_KUSTOMIZE {
		t.Errorf("expected KUSTOMIZE, got %v", got)
	}
}

func TestDetect_Plain(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/app.yaml", []byte("kind: Deployment"), 0o644); err != nil {
		t.Fatalf("writing app.yaml: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_PLAIN {
		t.Errorf("expected PLAIN, got %v", got)
	}
}

func TestDetect_Empty(t *testing.T) {
	fs := memfs.New()
	fs.MkdirAll("deploy", 0o755)

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_PLAIN {
		t.Errorf("expected PLAIN, got %v", got)
	}
}

func TestDetect_HelmTakesPriority(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}
	if err := util.WriteFile(fs, "deploy/kustomization.yaml", []byte("resources:"), 0o644); err != nil {
		t.Fatalf("writing kustomization.yaml: %v", err)
	}

	got := renderer.Detect(fs, "deploy")
	if got != keleustes.RendererType_RENDERER_TYPE_HELM {
		t.Errorf("expected HELM when both markers present, got %v", got)
	}
}
