package match

import (
	"testing"

	"github.com/google/go-github/v84/github"
	"github.com/katastroma/phortizo/internal/registration"
)

func pushEvent(ref, cloneURL string, commits []*github.HeadCommit) *github.PushEvent {
	return &github.PushEvent{
		Ref: new(ref),
		Repo: &github.PushEventRepository{
			CloneURL: new(cloneURL),
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

func TestMatches_RepoRefAndPath(t *testing.T) {
	wt := registration.WatchTarget{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if !matches(wt, "https://github.com/acme/app.git", "refs/heads/main", []string{"deploy/values.yaml"}) {
		t.Fatal("expected match when repo, ref, and path align")
	}
}

func TestMatches_DifferentRepo(t *testing.T) {
	wt := registration.WatchTarget{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if matches(wt, "https://github.com/other/repo.git", "refs/heads/main", []string{"deploy/values.yaml"}) {
		t.Fatal("expected no match for different repo")
	}
}

func TestMatches_DifferentRef(t *testing.T) {
	wt := registration.WatchTarget{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if matches(wt, "https://github.com/acme/app.git", "refs/heads/develop", []string{"deploy/values.yaml"}) {
		t.Fatal("expected no match for different ref")
	}
}

func TestMatches_OutsidePath(t *testing.T) {
	wt := registration.WatchTarget{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	if matches(wt, "https://github.com/acme/app.git", "refs/heads/main", []string{"src/main.go"}) {
		t.Fatal("expected no match for paths outside watched path")
	}
}

func TestChangedPaths_Deduplicates(t *testing.T) {
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, []string{"deploy/old.yaml"}, []string{"deploy/values.yaml"}),
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	paths := changedPaths(ev)

	seen := make(map[string]struct{})
	for _, p := range paths {
		if _, dup := seen[p]; dup {
			t.Fatalf("duplicate path in result: %s", p)
		}
		seen[p] = struct{}{}
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 unique paths, got %d", len(paths))
	}
}

func TestFind_MultipleMatches(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/prod/"},
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/staging/"},
	}
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/prod/values.yaml", "deploy/staging/values.yaml"}, nil, nil),
	})

	matched := Find(targets, ev)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matched))
	}
}

func TestFind_EmptyTargets(t *testing.T) {
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	if len(Find(nil, ev)) != 0 {
		t.Fatal("expected no matches for nil targets")
	}
}
