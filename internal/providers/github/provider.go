package github

import (
	"context"

	"github.com/joelklabo/ceye/internal/core"
)

// GitHubClient defines the subset of GitHub API operations this provider needs.
type GitHubClient interface {
	ListWorkflowRuns(owner, repo string) ([]core.Run, error)
}

// RepoConfig describes one repository to monitor.
type RepoConfig struct {
	Owner string
	Repo  string
}

// Provider polls GitHub Actions workflow runs for configured repos.
type Provider struct {
	client GitHubClient
	repos  []RepoConfig
}

// NewProvider constructs a GitHub provider with the supplied client and repo list.
func NewProvider(client GitHubClient, repos []RepoConfig) *Provider {
	return &Provider{client: client, repos: repos}
}

// Name implements core.Provider.
func (p *Provider) Name() string {
	return "github"
}

// Start begins polling GitHub for run data. Implementation will be added in Step 10.
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	<-ctx.Done()
	return ctx.Err()
}
