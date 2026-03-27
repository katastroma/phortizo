package source_test

import (
	"bytes"
	"testing"

	"github.com/katastroma/phortizo/internal/source"
)

func TestWatchTargetFromConfigMap(t *testing.T) {
	data := map[string]string{
		"repo-url": "https://github.com/acme/app.git",
		"ref":      "refs/heads/main",
		"path":     "deploy/",
	}

	target, err := source.TargetFromConfigMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", target.RepoURL, "https://github.com/acme/app.git")
	}

	if target.Ref != "refs/heads/main" {
		t.Errorf("Ref = %q, want %q", target.Ref, "refs/heads/main")
	}

	if target.Path != "deploy/" {
		t.Errorf("Path = %q, want %q", target.Path, "deploy/")
	}

	if target.Overrides != nil {
		t.Errorf("Overrides = %q, want nil", target.Overrides)
	}
}

func TestWatchTargetFromConfigMap_WithOverrides(t *testing.T) {
	overrides := "image:\n  tag: v1.2.3\n"
	data := map[string]string{
		"repo-url":  "https://github.com/acme/app.git",
		"ref":       "refs/heads/main",
		"path":      "deploy/",
		"overrides": overrides,
	}

	target, err := source.TargetFromConfigMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(target.Overrides) != overrides {
		t.Errorf("Overrides = %q, want %q", target.Overrides, overrides)
	}
}

func TestWatchTargetFromConfigMap_MissingRepoURL(t *testing.T) {
	data := map[string]string{"ref": "refs/heads/main", "path": "deploy/"}

	_, err := source.TargetFromConfigMap(data)
	if err == nil {
		t.Fatal("expected error for missing repo-url")
	}
}

func TestWatchTargetFromConfigMap_MissingRef(t *testing.T) {
	data := map[string]string{"repo-url": "https://github.com/acme/app.git", "path": "deploy/"}

	_, err := source.TargetFromConfigMap(data)
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
}

func TestWatchTargetFromConfigMap_MissingPath(t *testing.T) {
	data := map[string]string{"repo-url": "https://github.com/acme/app.git", "ref": "refs/heads/main"}

	_, err := source.TargetFromConfigMap(data)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestMarshalConfigMap(t *testing.T) {
	target := source.Target{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	data := target.MarshalConfigMap()

	if data["repo-url"] != "https://github.com/acme/app.git" {
		t.Errorf("repo-url = %q, want %q", data["repo-url"], "https://github.com/acme/app.git")
	}

	if data["ref"] != "refs/heads/main" {
		t.Errorf("ref = %q, want %q", data["ref"], "refs/heads/main")
	}

	if data["path"] != "deploy/" {
		t.Errorf("path = %q, want %q", data["path"], "deploy/")
	}

	if _, ok := data["overrides"]; ok {
		t.Error("expected no overrides key when Overrides is empty")
	}
}

func TestMarshalConfigMap_WithOverrides(t *testing.T) {
	overrides := []byte("image:\n  tag: v1.2.3\n")
	target := source.Target{
		RepoURL:   "https://github.com/acme/app.git",
		Ref:       "refs/heads/main",
		Path:      "deploy/",
		Overrides: overrides,
	}

	data := target.MarshalConfigMap()

	if data["overrides"] != string(overrides) {
		t.Errorf("overrides = %q, want %q", data["overrides"], overrides)
	}
}

func TestMarshalConfigMap_RoundTrip(t *testing.T) {
	original := source.Target{
		RepoURL:   "https://github.com/acme/app.git",
		Ref:       "refs/heads/main",
		Path:      "deploy/",
		Overrides: []byte("replicas: 3\n"),
	}

	data := original.MarshalConfigMap()

	restored, err := source.TargetFromConfigMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.RepoURL != original.RepoURL {
		t.Errorf("RepoURL = %q, want %q", restored.RepoURL, original.RepoURL)
	}

	if restored.Ref != original.Ref {
		t.Errorf("Ref = %q, want %q", restored.Ref, original.Ref)
	}

	if restored.Path != original.Path {
		t.Errorf("Path = %q, want %q", restored.Path, original.Path)
	}

	if !bytes.Equal(restored.Overrides, original.Overrides) {
		t.Errorf("Overrides = %q, want %q", restored.Overrides, original.Overrides)
	}
}
