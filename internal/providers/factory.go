package providers

import (
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	azureprovider "github.com/joelklabo/ceye/internal/providers/azure"
	demoprovider "github.com/joelklabo/ceye/internal/providers/demo"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
	gitlabprovider "github.com/joelklabo/ceye/internal/providers/gitlab"
)

// ProviderConfig is a generic configuration for any provider entry.
type ProviderConfig struct {
	Type string `mapstructure:"type"`

	// GitHub-specific fields
	Repos []githubprovider.RepoConfig `mapstructure:"repos"`

	// Azure-specific fields
	Org       string `mapstructure:"org"`
	Project   string `mapstructure:"project"`
	Pipelines []int  `mapstructure:"pipelines"`

	// Demo-specific fields
	Runs int `mapstructure:"runs"`

	// GitLab-specific fields
	GitLabProject string `mapstructure:"gitlab_project"`
}

// Dependencies supplies shared resources (API clients, tokens, etc.) needed
// when constructing concrete provider instances.
type Dependencies struct {
	GitHubClient githubprovider.GitHubClient
	AzureClient  azureprovider.AzureClient
}

// CreateProvider instantiates a concrete provider implementation based on cfg.
func CreateProvider(cfg ProviderConfig, deps Dependencies) (core.Provider, error) {
	switch cfg.Type {
	case "github":
		if deps.GitHubClient == nil {
			return nil, fmt.Errorf("github client is required")
		}
		if len(cfg.Repos) == 0 {
			return nil, fmt.Errorf("github provider requires at least one repo")
		}
		return githubprovider.NewProvider(deps.GitHubClient, cfg.Repos), nil
	case "azure":
		if deps.AzureClient == nil {
			return nil, fmt.Errorf("azure client is required")
		}
		if cfg.Org == "" || cfg.Project == "" {
			return nil, fmt.Errorf("azure provider requires org and project")
		}
		azureCfg := azureprovider.Config{Org: cfg.Org, Project: cfg.Project, Pipelines: cfg.Pipelines}
		return azureprovider.NewProvider(deps.AzureClient, azureCfg), nil
	case "demo":
		count := cfg.Runs
		return demoprovider.New(count, 5*time.Second), nil
	case "gitlab":
		if cfg.GitLabProject == "" {
			return nil, fmt.Errorf("gitlab provider requires project")
		}
		return gitlabprovider.NewProvider(gitlabprovider.Config{Project: cfg.GitLabProject}), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}
