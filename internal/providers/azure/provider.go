package azure

import (
	"context"
	"log"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// AzureClient abstracts Azure DevOps API interactions used by the provider.
type AzureClient interface {
	ListBuilds(org, project string, pipelines []int) ([]core.Run, error)
}

// Config describes the Azure DevOps project/pipeline selection.
type Config struct {
	Org       string
	Project   string
	Pipelines []int
}

const (
	azureFastInterval = 15 * time.Second
	azureSlowInterval = 60 * time.Second
)

// Provider polls Azure DevOps builds for configured pipelines.
type Provider struct {
	client       AzureClient
	cfg          Config
	fastInterval time.Duration
	slowInterval time.Duration
	refreshCh    chan struct{}
	lastError    error
}

// NewProvider constructs an Azure provider.
func NewProvider(client AzureClient, cfg Config) *Provider {
	return &Provider{
		client:       client,
		cfg:          cfg,
		fastInterval: azureFastInterval,
		slowInterval: azureSlowInterval,
		refreshCh:    make(chan struct{}, 1),
	}
}

// Name implements core.Provider.
func (p *Provider) Name() string {
	return "azure"
}

// Start begins polling Azure for builds.
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	interval := p.fastInterval

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		runs, err := p.client.ListBuilds(p.cfg.Org, p.cfg.Project, p.cfg.Pipelines)
		if err != nil {
			p.lastError = err
			log.Printf("azure provider error for %s/%s: %v", p.cfg.Org, p.cfg.Project, err)
			p.emitEvent(ctx, out, nil, err)
		} else if len(runs) > 0 {
			p.lastError = nil
			p.emitEvent(ctx, out, runs, nil)
		}

		interval = p.nextInterval(runs, interval)

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
