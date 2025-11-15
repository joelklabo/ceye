package providers

import (
	"fmt"
	"strings"
)

// DisplayName returns the friendly label for a provider config.
func DisplayName(cfg ProviderConfig) string {
	if name := strings.TrimSpace(cfg.DisplayName); name != "" {
		return name
	}
	if cfg.Type != "" {
		return fmt.Sprintf("%s provider", cfg.Type)
	}
	return "provider"
}

// StoreDetail renders a detail string for the overlay.
func StoreDetail(cfg ProviderConfig) string {
	switch cfg.Type {
	case "github":
		if len(cfg.Repos) == 0 {
			return "GitHub (repos unspecified)"
		}
		parts := make([]string, 0, len(cfg.Repos))
		for _, repo := range cfg.Repos {
			parts = append(parts, fmt.Sprintf("%s/%s", repo.Owner, repo.Repo))
		}
		return fmt.Sprintf("GitHub repos: %s", strings.Join(parts, ", "))
	case "azure":
		if cfg.Org == "" || cfg.Project == "" {
			return "Azure (org/project missing)"
		}
		if len(cfg.Pipelines) == 0 {
			return fmt.Sprintf("Azure %s/%s (all pipelines)", cfg.Org, cfg.Project)
		}
		pipes := make([]string, len(cfg.Pipelines))
		for i, id := range cfg.Pipelines {
			pipes[i] = fmt.Sprintf("%d", id)
		}
		return fmt.Sprintf("Azure %s/%s pipelines: %s", cfg.Org, cfg.Project, strings.Join(pipes, ", "))
	case "demo":
		if cfg.Runs > 0 {
			return fmt.Sprintf("Demo runs: %d", cfg.Runs)
		}
		return "Demo provider"
	case "gitlab":
		if cfg.GitLabProject != "" {
			return fmt.Sprintf("GitLab project: %s", cfg.GitLabProject)
		}
		return "GitLab provider"
	default:
		return ""
	}
}
