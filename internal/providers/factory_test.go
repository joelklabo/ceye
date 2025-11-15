package providers

import (
	"testing"

	"github.com/joelklabo/ceye/internal/core"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
)

type stubGitHubClient struct{}

func (s stubGitHubClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	return nil, nil
}

type stubAzureClient struct{}

func (s stubAzureClient) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
	return nil, nil
}

func TestCreateProvider_GitHub(t *testing.T) {
	cfg := ProviderConfig{
		Type: "github",
		Repos: []githubprovider.RepoConfig{
			{Owner: "octocat", Repo: "hello"},
		},
	}
	provider, err := CreateProvider(cfg, Dependencies{GitHubClient: stubGitHubClient{}})
	if err != nil {
		t.Fatalf("expected github provider, got error: %v", err)
	}
	if provider.Name() != "github" {
		t.Fatalf("expected github provider name, got %s", provider.Name())
	}
}

func TestCreateProvider_Azure(t *testing.T) {
	cfg := ProviderConfig{
		Type:      "azure",
		Org:       "org",
		Project:   "project",
		Pipelines: []int{1, 2},
	}
	provider, err := CreateProvider(cfg, Dependencies{AzureClient: stubAzureClient{}})
	if err != nil {
		t.Fatalf("expected azure provider, got error: %v", err)
	}
	if provider.Name() != "azure" {
		t.Fatalf("expected azure provider name, got %s", provider.Name())
	}
}

func TestCreateProviderUnknownType(t *testing.T) {
	_, err := CreateProvider(ProviderConfig{Type: "gitlab"}, Dependencies{})
	if err == nil {
		t.Fatalf("expected error for unknown provider type")
	}
}
