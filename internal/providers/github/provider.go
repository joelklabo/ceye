package github

import (
	"context"
	"log"
	"time"

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

const (
	defaultFastInterval = 15 * time.Second
	defaultSlowInterval = 60 * time.Second
)

// Provider polls GitHub Actions workflow runs for configured repos.
type Provider struct {
	client       GitHubClient
	repos        []RepoConfig
	fastInterval time.Duration
	slowInterval time.Duration
	refreshCh    chan struct{}
	lastError    error
}

// NewProvider constructs a GitHub provider with the supplied client and repo list.
func NewProvider(client GitHubClient, repos []RepoConfig) *Provider {
	return &Provider{
		client:       client,
		repos:        repos,
		fastInterval: defaultFastInterval,
		slowInterval: defaultSlowInterval,
		refreshCh:    make(chan struct{}, 1),
	}
}

// Name implements core.Provider.
func (p *Provider) Name() string {
	return "github"
}

// Start begins polling GitHub for run data.
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	if len(p.repos) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}

	interval := p.fastInterval

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var combined []core.Run
		for _, repo := range p.repos {
			runs, err := p.client.ListWorkflowRuns(repo.Owner, repo.Repo)
			if err != nil {
				p.lastError = err
				log.Printf("github provider error for %s/%s: %v", repo.Owner, repo.Repo, err)
				p.emitEvent(ctx, out, nil, err)
				continue
			}
			combined = append(combined, runs...)
		}

		if len(combined) > 0 {
			p.lastError = nil
			p.emitEvent(ctx, out, combined, nil)
		}

		interval = p.nextInterval(combined, interval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.refreshCh:
			continue
		case <-time.After(interval):
		}
	}
}

func (p *Provider) nextInterval(runs []core.Run, current time.Duration) time.Duration {
	if hasActiveRuns(runs) {
		return p.fastInterval
	}
	if current < p.slowInterval {
		return p.slowInterval
	}
	return current
}

func hasActiveRuns(runs []core.Run) bool {
	for _, run := range runs {
		switch run.Status {
		case core.RunStatusInProgress, core.RunStatusQueued:
			return true
		}
	}
	return false
}

// TriggerRefresh requests an immediate polling cycle.
func (p *Provider) TriggerRefresh() {
	select {
	case p.refreshCh <- struct{}{}:
	default:
	}
}

// LastError returns the last polling error encountered.
func (p *Provider) LastError() error {
	return p.lastError
}

func (p *Provider) emitEvent(ctx context.Context, out chan<- core.RunEvent, runs []core.Run, err error) {
	select {
	case out <- core.RunEvent{Provider: p.Name(), Runs: runs, Timestamp: time.Now(), Err: err}:
	case <-ctx.Done():
	}
}
