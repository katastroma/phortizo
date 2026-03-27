package renderer

import (
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

func createFile(t *testing.T, fs billy.Filesystem, path string) {
	t.Helper()
	f, err := fs.Create(path)
	if err != nil {
		t.Fatalf("creating %s: %v", path, err)
	}
	f.Close()
}

func TestDetectRenderer_Helm(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/Chart.yaml")

	if got := Detect(fs, "deploy"); got != Helm {
		t.Errorf("got %q, want %q", got, Helm)
	}
}

func TestDetectRenderer_KustomizeYaml(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/kustomization.yaml")

	if got := Detect(fs, "deploy"); got != Kustomize {
		t.Errorf("got %q, want %q", got, Kustomize)
	}
}

func TestDetectRenderer_KustomizeYml(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/kustomization.yml")

	if got := Detect(fs, "deploy"); got != Kustomize {
		t.Errorf("got %q, want %q", got, Kustomize)
	}
}

func TestDetectRenderer_KustomizeCapitalized(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/Kustomization")

	if got := Detect(fs, "deploy"); got != Kustomize {
		t.Errorf("got %q, want %q", got, Kustomize)
	}
}

func TestDetectRenderer_Raw(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/deployment.yaml")

	if got := Detect(fs, "deploy"); got != Raw {
		t.Errorf("got %q, want %q", got, Raw)
	}
}

func TestDetectRenderer_NoMarkerFiles(t *testing.T) {
	fs := memfs.New()

	if got := Detect(fs, "nonexistent"); got != Raw {
		t.Errorf("got %q, want %q", got, Raw)
	}
}

func TestDetectRenderer_HelmTakesPrecedence(t *testing.T) {
	fs := memfs.New()
	createFile(t, fs, "deploy/Chart.yaml")
	createFile(t, fs, "deploy/kustomization.yaml")

	if got := Detect(fs, "deploy"); got != Helm {
		t.Errorf("got %q, want %q (Helm should take precedence)", got, Helm)
	}
}
