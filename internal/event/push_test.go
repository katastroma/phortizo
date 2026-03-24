package event

import (
	"sort"
	"testing"
)

func TestParsePush_ValidPayload(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [
			{"added": ["deploy/values.yaml"], "removed": [], "modified": ["README.md"]},
			{"added": [], "removed": ["old.txt"], "modified": ["deploy/values.yaml"]}
		]
	}`)

	ev, err := ParsePush(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.Ref != "refs/heads/main" {
		t.Errorf("ref = %q, want %q", ev.Ref, "refs/heads/main")
	}
	if ev.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("repo_url = %q, want %q", ev.RepoURL, "https://github.com/acme/app.git")
	}

	sort.Strings(ev.ChangedPaths)
	want := []string{"README.md", "deploy/values.yaml", "old.txt"}
	if len(ev.ChangedPaths) != len(want) {
		t.Fatalf("changed_paths length = %d, want %d", len(ev.ChangedPaths), len(want))
	}
	for i, p := range ev.ChangedPaths {
		if p != want[i] {
			t.Errorf("changed_paths[%d] = %q, want %q", i, p, want[i])
		}
	}
}

func TestParsePush_DeduplicatesPaths(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [
			{"added": ["file.txt"], "removed": [], "modified": ["file.txt"]},
			{"added": ["file.txt"], "removed": [], "modified": []}
		]
	}`)

	ev, err := ParsePush(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ev.ChangedPaths) != 1 {
		t.Errorf("expected 1 deduplicated path, got %d: %v", len(ev.ChangedPaths), ev.ChangedPaths)
	}
}

func TestParsePush_InvalidJSON(t *testing.T) {
	if _, err := ParsePush([]byte("not json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePush_EmptyCommits(t *testing.T) {
	body := []byte(`{
		"ref": "refs/heads/main",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": []
	}`)

	ev, err := ParsePush(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ev.ChangedPaths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(ev.ChangedPaths))
	}
}
