package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverGitRepos(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "repo", ".git"), 0o755)
	repos, err := discoverGitRepos(root)
	if err != nil {
		t.Fatalf("discover git repos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
}

func TestWorkspaceRootBase(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/Users/honk/code/ceye", "/Users/honk/code"},
		{"/Users/honk/code/sub/dir", "/Users/honk/code"},
		{"/Users/honk/project", "/Users/honk/project"},
		{"/code/ceye", "/code"},
	}
	for _, tt := range tests {
		if got := workspaceRootBase(tt.path); got != tt.want {
			t.Fatalf("workspaceRootBase(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestResolveConfigRoot(t *testing.T) {
	os.Setenv(envConfigRoot, "/foo")
	defer os.Unsetenv(envConfigRoot)
	if got, _ := resolveConfigRoot(""); got != "/foo" {
		t.Fatalf("expected override from env, got %s", got)
	}
	if _, err := resolveConfigRoot("/bar"); err != nil {
		t.Fatalf("expected no error when flag provided: %v", err)
	}
	os.Unsetenv(envConfigRoot)
	if got, _ := resolveConfigRoot("."); got == "" {
		t.Fatalf("expected fallback to workspace when no override")
	}
}

func TestListMissingConfigs(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git repo: %v", err)
	}
	missing, err := listMissingConfigs(root)
	if err != nil {
		t.Fatalf("list missing configs: %v", err)
	}
	if len(missing) != 1 {
		t.Fatalf("expected 1 missing repo, got %d", len(missing))
	}
	if err := os.WriteFile(filepath.Join(repo, "ceye.yaml"), []byte("providers: []\n"), 0o644); err != nil {
		t.Fatalf("create config: %v", err)
	}
	updated, err := listMissingConfigs(root)
	if err != nil {
		t.Fatalf("list missing configs after create: %v", err)
	}
	if len(updated) != 0 {
		t.Fatalf("expected no missing configs after creation, got %v", updated)
	}
}
