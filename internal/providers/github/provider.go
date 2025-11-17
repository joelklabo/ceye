package github

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	webhookMode  bool // If true, do one initial poll then wait for webhooks
}

// NewProvider constructs a GitHub provider with the supplied client and repo list.
func NewProvider(client GitHubClient, repos []RepoConfig) *Provider {
	return &Provider{
		client:       client,
		repos:        repos,
		fastInterval: defaultFastInterval,
		slowInterval: defaultSlowInterval,
		refreshCh:    make(chan struct{}, 1),
		webhookMode:  false,
	}
}

// SetWebhookMode enables webhook-only mode where the provider does one initial
// poll to load data, then waits for webhook updates instead of continuous polling.
func (p *Provider) SetWebhookMode(enabled bool) {
	p.webhookMode = enabled
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
	pollCount := 0
	
	// In webhook mode: do one initial poll, then wait for context cancellation
	if p.webhookMode {
		log.Printf("github: webhook mode enabled - doing initial poll only")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pollCount++
		
		// In webhook mode, only do the first poll
		if p.webhookMode && pollCount > 1 {
			log.Printf("github: webhook mode - initial poll complete, waiting for webhooks")
			<-ctx.Done()
			return ctx.Err()
		}
		
		log.Printf("github: poll cycle #%d starting (%d repos, interval %v)", pollCount, len(p.repos), interval)

		var combined []core.Run
		rateLimitHit := false
		for _, repo := range p.repos {
			runs, err := p.client.ListWorkflowRuns(repo.Owner, repo.Repo)
			if err != nil {
				// Check for rate limiting
				errStr := err.Error()
				if strings.Contains(errStr, "rate limit") {
					if pollCount == 1 || pollCount%10 == 0 {
						log.Printf("github provider: ⚠️  RATE LIMIT EXCEEDED - will retry with backoff (cycle %d)", pollCount)
					}
					p.lastError = fmt.Errorf("rate limit exceeded")
					p.emitEvent(ctx, out, nil, err)
					rateLimitHit = true
					break // Stop trying other repos this cycle
				} else if strings.Contains(errStr, "exit status 1") {
					// Silently skip repos without workflows - this is expected
					continue
				} else {
					log.Printf("github provider error for %s/%s: %v", repo.Owner, repo.Repo, err)
					p.lastError = err
					p.emitEvent(ctx, out, nil, err)
				}
				continue
			}
			combined = append(combined, runs...)
		}

		log.Printf("github: poll cycle #%d complete - fetched %d runs", pollCount, len(combined))

		if len(combined) > 0 {
			p.lastError = nil
			p.emitEvent(ctx, out, combined, nil)
		}

		// Handle rate limit backoff
		if rateLimitHit {
			interval = p.slowInterval * 4 // 240s backoff when rate limited
			log.Printf("github: backing off to %v due to rate limit", interval)
		} else {
			interval = p.nextInterval(combined, interval)
		}

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
