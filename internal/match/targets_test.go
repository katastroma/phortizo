package match

import (
	"testing"

	"github.com/katastroma/phortizo/internal/event"
	"github.com/katastroma/phortizo/internal/registration"
)

func TestTargets_MatchesRepoRefAndPath(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := event.Push{
		RepoURL:      "https://github.com/acme/app.git",
		Ref:          "refs/heads/main",
		ChangedPaths: []string{"deploy/values.yaml"},
	}

	matched := Targets(targets, ev)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
}

func TestTargets_NoMatchDifferentRepo(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := event.Push{
		RepoURL:      "https://github.com/other/repo.git",
		Ref:          "refs/heads/main",
		ChangedPaths: []string{"deploy/values.yaml"},
	}

	if len(Targets(targets, ev)) != 0 {
		t.Fatal("expected no match for different repo")
	}
}

func TestTargets_NoMatchDifferentRef(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := event.Push{
		RepoURL:      "https://github.com/acme/app.git",
		Ref:          "refs/heads/develop",
		ChangedPaths: []string{"deploy/values.yaml"},
	}

	if len(Targets(targets, ev)) != 0 {
		t.Fatal("expected no match for different ref")
	}
}

func TestTargets_NoMatchOutsidePath(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/"},
	}
	ev := event.Push{
		RepoURL:      "https://github.com/acme/app.git",
		Ref:          "refs/heads/main",
		ChangedPaths: []string{"src/main.go"},
	}

	if len(Targets(targets, ev)) != 0 {
		t.Fatal("expected no match for paths outside watched path")
	}
}

func TestTargets_MultipleMatches(t *testing.T) {
	targets := []registration.WatchTarget{
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/prod/"},
		{RepoURL: "https://github.com/acme/app.git", Ref: "refs/heads/main", Path: "deploy/staging/"},
	}
	ev := event.Push{
		RepoURL:      "https://github.com/acme/app.git",
		Ref:          "refs/heads/main",
		ChangedPaths: []string{"deploy/prod/values.yaml", "deploy/staging/values.yaml"},
	}

	matched := Targets(targets, ev)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matched))
	}
}

func TestTargets_EmptyTargets(t *testing.T) {
	ev := event.Push{
		RepoURL:      "https://github.com/acme/app.git",
		Ref:          "refs/heads/main",
		ChangedPaths: []string{"deploy/values.yaml"},
	}

	if len(Targets(nil, ev)) != 0 {
		t.Fatal("expected no matches for nil targets")
	}
}
