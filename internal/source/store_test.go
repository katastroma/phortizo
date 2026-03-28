//revive:disable:package-comments
package source_test

import (
	"fmt"
	"testing"

	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/source"
	"github.com/katastroma/phortizo/internal/tests"
)

func seedTarget(store *tests.MockStore, name string) {
	obj := tests.NewMockObject(map[string]string{
		source.CredentialSecretAnnotation: "my-cred",
	})
	obj.SetName(name)
	obj.SetLabels(map[string]string{object.TypeLabel: source.TypeLabel})
	obj.SetData(map[string]string{
		"repo-url": "https://github.com/acme/app.git",
		"ref":      "refs/heads/main",
		"path":     "deploy/",
	})
	store.Add(name, obj)
}

func TestTargetFromResource(t *testing.T) {
	obj := tests.NewMockObject(map[string]string{
		source.CredentialSecretAnnotation: "my-cred",
	})
	obj.SetName("wt-1")
	obj.SetData(map[string]string{
		"repo-url": "https://github.com/acme/app.git",
		"ref":      "refs/heads/main",
		"path":     "deploy/",
	})

	target, err := source.TargetFromResource(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Name != "wt-1" {
		t.Errorf("Name = %q, want %q", target.Name, "wt-1")
	}

	if target.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", target.RepoURL, "https://github.com/acme/app.git")
	}

	if target.CredentialSecret != "my-cred" {
		t.Errorf("CredentialSecret = %q, want %q", target.CredentialSecret, "my-cred")
	}
}

func TestTargetFromResource_MissingData(t *testing.T) {
	obj := tests.NewMockObject(nil)
	obj.SetData(map[string]string{"ref": "refs/heads/main"})

	_, err := source.TargetFromResource(obj)
	if err == nil {
		t.Fatal("expected error for missing repo-url")
	}
}

func TestGet(t *testing.T) {
	store := tests.NewMockStore()
	seedTarget(store, "wt-1")

	target, err := source.Get(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Name != "wt-1" {
		t.Errorf("Name = %q, want %q", target.Name, "wt-1")
	}
}

func TestGet_NotFound(t *testing.T) {
	store := tests.NewMockStore()

	_, err := source.Get(t.Context(), store, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing resource")
	}
}

func TestList(t *testing.T) {
	store := tests.NewMockStore()
	seedTarget(store, "wt-1")
	seedTarget(store, "wt-2")

	targets, err := source.List(t.Context(), store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestList_Empty(t *testing.T) {
	store := tests.NewMockStore()

	targets, err := source.List(t.Context(), store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(targets))
	}
}

func TestList_Error(t *testing.T) {
	store := tests.NewMockStore()
	store.ListErr = fmt.Errorf("list denied")

	_, err := source.List(t.Context(), store)
	if err == nil {
		t.Fatal("expected error from failing list")
	}
}

func TestPut_Create(t *testing.T) {
	store := tests.NewMockStore()

	target := &source.Target{
		Name:             "wt-1",
		RepoURL:          "https://github.com/acme/app.git",
		Ref:              "refs/heads/main",
		Path:             "deploy/",
		CredentialSecret: "my-cred",
	}

	if err := source.Put(t.Context(), store, target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := source.Get(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if got.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", got.RepoURL, "https://github.com/acme/app.git")
	}

	if got.CredentialSecret != "my-cred" {
		t.Errorf("CredentialSecret = %q, want %q", got.CredentialSecret, "my-cred")
	}
}

func TestPut_Update(t *testing.T) {
	store := tests.NewMockStore()
	seedTarget(store, "wt-1")

	target := &source.Target{
		Name:    "wt-1",
		RepoURL: "https://github.com/acme/new.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/prod/",
	}

	if err := source.Put(t.Context(), store, target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := source.Get(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if got.RepoURL != "https://github.com/acme/new.git" {
		t.Errorf("RepoURL = %q, want %q", got.RepoURL, "https://github.com/acme/new.git")
	}

	if got.CredentialSecret != "" {
		t.Errorf("CredentialSecret = %q, want empty", got.CredentialSecret)
	}
}

func TestPut_CreateNoCredential(t *testing.T) {
	store := tests.NewMockStore()

	target := &source.Target{
		Name:    "wt-1",
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if err := source.Put(t.Context(), store, target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := source.Get(t.Context(), store, "wt-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	if got.CredentialSecret != "" {
		t.Errorf("CredentialSecret = %q, want empty", got.CredentialSecret)
	}
}

func TestPut_StoreError(t *testing.T) {
	store := tests.NewMockStore()
	store.PutErr = fmt.Errorf("put denied")

	target := &source.Target{
		Name:    "wt-1",
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if err := source.Put(t.Context(), store, target); err == nil {
		t.Fatal("expected error from failing put")
	}
}
