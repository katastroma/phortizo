package event

import (
	"sort"
	"testing"

	"github.com/google/go-github/v84/github"
)

const testCommitSHA = "abc123def456"

func pushEvent(ref, cloneURL string, commits []*github.HeadCommit) *github.PushEvent {
	return &github.PushEvent{
		Ref:   github.Ptr(ref),
		After: github.Ptr(testCommitSHA),
		Repo: &github.PushEventRepository{
			CloneURL: github.Ptr(cloneURL),
		},
		Commits: commits,
	}
}

func commit(added, removed, modified []string) *github.HeadCommit {
	return &github.HeadCommit{
		Added:    added,
		Removed:  removed,
		Modified: modified,
	}
}

func TestFromPushEvent_ValidPayload(t *testing.T) {
	ev := FromPushEvent(pushEvent(
		"refs/heads/main",
		"https://github.com/acme/app.git",
		[]*github.HeadCommit{
			commit([]string{"deploy/values.yaml"}, nil, []string{"README.md"}),
			commit(nil, []string{"old.txt"}, []string{"deploy/values.yaml"}),
		},
	))

	if ev.Ref != "refs/heads/main" {
		t.Errorf("ref = %q, want %q", ev.Ref, "refs/heads/main")
	}
	if ev.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("repo_url = %q, want %q", ev.RepoURL, "https://github.com/acme/app.git")
	}
	if ev.CommitSHA != testCommitSHA {
		t.Errorf("commit_sha = %q, want %q", ev.CommitSHA, testCommitSHA)
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

func TestFromPushEvent_DeduplicatesPaths(t *testing.T) {
	ev := FromPushEvent(pushEvent(
		"refs/heads/main",
		"https://github.com/acme/app.git",
		[]*github.HeadCommit{
			commit([]string{"file.txt"}, nil, []string{"file.txt"}),
			commit([]string{"file.txt"}, nil, nil),
		},
	))

	if len(ev.ChangedPaths) != 1 {
		t.Errorf("expected 1 deduplicated path, got %d: %v", len(ev.ChangedPaths), ev.ChangedPaths)
	}
}

func TestFromPushEvent_EmptyCommits(t *testing.T) {
	ev := FromPushEvent(pushEvent(
		"refs/heads/main",
		"https://github.com/acme/app.git",
		nil,
	))

	if len(ev.ChangedPaths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(ev.ChangedPaths))
	}
}
