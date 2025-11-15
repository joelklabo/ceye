package providers

import (
	"strings"
	"testing"

	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
)

func TestStoreDetail(t *testing.T) {
	if got := StoreDetail(ProviderConfig{Type: "github", Repos: []githubprovider.RepoConfig{{Owner: "o", Repo: "r"}}}); !contains(got, "o/r") {
		t.Fatalf("expected github repo detail, got %q", got)
	}
	if got := StoreDetail(ProviderConfig{Type: "azure", Org: "org", Project: "proj", Pipelines: []int{1, 2}}); !contains(got, "org/proj") {
		t.Fatalf("expected azure detail, got %q", got)
	}
	if got := StoreDetail(ProviderConfig{Type: "demo", Runs: 3}); got != "Demo runs: 3" {
		t.Fatalf("expected demo detail, got %q", got)
	}
	if got := StoreDetail(ProviderConfig{Type: "gitlab", GitLabProject: "p"}); !contains(got, "GitLab") {
		t.Fatalf("expected gitlab detail, got %q", got)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
