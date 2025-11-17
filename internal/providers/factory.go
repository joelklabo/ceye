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
	// DisplayName optionally overrides the provider label shown in the UI.
	DisplayName string `mapstructure:"display_name"`
	// Logo is an optional path to a custom SVG logo (e.g. "/logos/jenkins.svg")
	// Logo should be SVG with 24x24 viewBox for best results
	Logo string `mapstructure:"logo"`

	// GitHub-specific fields
	Repos []githubprovider.RepoConfig `mapstructure:"repos"`

	// Azure-specific fields
	Org          string                        `mapstructure:"org"`
	Project      string                        `mapstructure:"project"`      // Deprecated: use Projects
	Pipelines    []int                         `mapstructure:"pipelines"`    // Deprecated: use Projects
	Projects     []azureprovider.ProjectConfig `mapstructure:"projects"`     // Multi-project support
	FastInterval string                        `mapstructure:"fast_interval"` // e.g. "15s"
	SlowInterval string                        `mapstructure:"slow_interval"` // e.g. "60s"

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
		if cfg.Org == "" {
			return nil, fmt.Errorf("azure provider requires org")
		}
		
		// Build Azure config
		azureCfg := azureprovider.Config{
			DisplayName: cfg.DisplayName,
			Org:         cfg.Org,
		}
		
		// Support legacy single-project config or new multi-project
		if len(cfg.Projects) > 0 {
			azureCfg.Projects = cfg.Projects
		} else if cfg.Project != "" {
			// Legacy single-project config
			azureCfg.Projects = []azureprovider.ProjectConfig{
				{Name: cfg.Project, Pipelines: cfg.Pipelines},
			}
		} else {
			return nil, fmt.Errorf("azure provider requires projects or project")
		}
		
		// Parse intervals if provided
		if cfg.FastInterval != "" {
			d, err := time.ParseDuration(cfg.FastInterval)
			if err != nil {
				return nil, fmt.Errorf("invalid fast_interval: %w", err)
			}
			azureCfg.FastInterval = d
		}
		if cfg.SlowInterval != "" {
			d, err := time.ParseDuration(cfg.SlowInterval)
			if err != nil {
				return nil, fmt.Errorf("invalid slow_interval: %w", err)
			}
			azureCfg.SlowInterval = d
		}
		
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
