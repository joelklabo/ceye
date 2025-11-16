package providers

import (
	"testing"

	"github.com/joelklabo/ceye/internal/core"
	azureprovider "github.com/joelklabo/ceye/internal/providers/azure"
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
	if provider.Name() != "azure-org" {
		t.Fatalf("expected azure-org provider name, got %s", provider.Name())
	}
}

func TestCreateProvider_AzureMultiProject(t *testing.T) {
	cfg := ProviderConfig{
		Type:        "azure",
		DisplayName: "Azure Prod",
		Org:         "myorg",
		Projects: []azureprovider.ProjectConfig{
			{Name: "Frontend", Pipelines: []int{1, 2}},
			{Name: "Backend", Pipelines: []int{3, 4}},
		},
		FastInterval: "10s",
		SlowInterval: "120s",
	}
	provider, err := CreateProvider(cfg, Dependencies{AzureClient: stubAzureClient{}})
	if err != nil {
		t.Fatalf("expected azure provider, got error: %v", err)
	}
	if provider.Name() != "Azure Prod" {
		t.Fatalf("expected 'Azure Prod' name, got %s", provider.Name())
	}
}

func TestCreateProvider_AzureInvalidInterval(t *testing.T) {
	cfg := ProviderConfig{
		Type:         "azure",
		Org:          "myorg",
		Project:      "proj",
		FastInterval: "invalid",
	}
	_, err := CreateProvider(cfg, Dependencies{AzureClient: stubAzureClient{}})
	if err == nil {
		t.Fatalf("expected error for invalid interval")
	}
}

func TestCreateProvider_AzureMissingOrg(t *testing.T) {
	cfg := ProviderConfig{
		Type:    "azure",
		Project: "proj",
	}
	_, err := CreateProvider(cfg, Dependencies{AzureClient: stubAzureClient{}})
	if err == nil {
		t.Fatalf("expected error for missing org")
	}
}

func TestCreateProvider_AzureMissingProjects(t *testing.T) {
	cfg := ProviderConfig{
		Type: "azure",
		Org:  "myorg",
	}
	_, err := CreateProvider(cfg, Dependencies{AzureClient: stubAzureClient{}})
	if err == nil {
		t.Fatalf("expected error for missing projects")
	}
}

func TestCreateProviderUnknownType(t *testing.T) {
	_, err := CreateProvider(ProviderConfig{Type: "gitlab"}, Dependencies{})
	if err == nil {
		t.Fatalf("expected error for unknown provider type")
	}
}
