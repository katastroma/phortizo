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

func TestTargets_MatchesRepoRefAndPath(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	matched := Find(targets, ev)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
}

func TestTargets_NoMatchDifferentRepo(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := pushEvent("refs/heads/main", "https://github.com/other/repo.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	if len(Find(targets, ev)) != 0 {
		t.Fatal("expected no match for different repo")
	}
}

func TestTargets_NoMatchDifferentRef(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := pushEvent("refs/heads/develop", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	if len(Find(targets, ev)) != 0 {
		t.Fatal("expected no match for different ref")
	}
}

func TestTargets_NoMatchOutsidePath(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit(nil, nil, []string{"src/main.go"}),
	})

	if len(Find(targets, ev)) != 0 {
		t.Fatal("expected no match for paths outside watched path")
	}
}

func TestTargets_MultipleMatches(t *testing.T) {
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

func TestTargets_EmptyTargets(t *testing.T) {
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	if len(Find(nil, ev)) != 0 {
		t.Fatal("expected no matches for nil targets")
	}
}

func TestTargets_DeduplicatesPaths(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := pushEvent("refs/heads/main", "https://github.com/acme/app.git", []*github.HeadCommit{
		commit([]string{"deploy/values.yaml"}, []string{"deploy/old.yaml"}, []string{"deploy/values.yaml"}),
		commit([]string{"deploy/values.yaml"}, nil, nil),
	})

	matched := Find(targets, ev)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
}
