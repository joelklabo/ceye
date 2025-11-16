package azure

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// AzureClient abstracts Azure DevOps API interactions used by the provider.
type AzureClient interface {
	ListBuilds(org, project string, pipelines []int) ([]core.Run, error)
}

// ProjectConfig describes a single Azure DevOps project and its pipelines
type ProjectConfig struct {
	Name      string
	Pipelines []int // Empty means all pipelines
}

// Config describes the Azure DevOps provider configuration
type Config struct {
	DisplayName string          // Optional friendly name for this provider
	Org         string          // Azure DevOps organization
	Projects    []ProjectConfig // Projects to monitor
	
	// Optional: Override default polling intervals
	FastInterval time.Duration
	SlowInterval time.Duration
}

const (
	azureFastInterval = 15 * time.Second
	azureSlowInterval = 60 * time.Second
)

// Provider polls Azure DevOps builds for configured projects.
type Provider struct {
	client       AzureClient
	cfg          Config
	name         string
	fastInterval time.Duration
	slowInterval time.Duration
	refreshCh    chan struct{}
	lastError    error
}

// NewProvider constructs an Azure provider with the given client and configuration.
func NewProvider(client AzureClient, cfg Config) *Provider {
	name := cfg.DisplayName
	if name == "" {
		name = "azure-" + cfg.Org
	}
	
	fastInterval := cfg.FastInterval
	if fastInterval == 0 {
		fastInterval = azureFastInterval
	}
	
	slowInterval := cfg.SlowInterval
	if slowInterval == 0 {
		slowInterval = azureSlowInterval
	}
	
	return &Provider{
		client:       client,
		cfg:          cfg,
		name:         name,
		fastInterval: fastInterval,
		slowInterval: slowInterval,
		refreshCh:    make(chan struct{}, 1),
	}
}

// NewProviderFromConfig creates a provider from configuration using environment variables.
func NewProviderFromConfig(cfg Config) *Provider {
	pat := os.Getenv("AZURE_PAT")
	if pat == "" {
		pat = os.Getenv("AZURE_DEVOPS_PAT")
	}
	
	client := NewClient(cfg.Org, pat)
	return NewProvider(client, cfg)
}

// Name implements core.Provider.
func (p *Provider) Name() string {
	return p.name
}

// Start begins polling Azure DevOps for builds across all configured projects.
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	interval := p.fastInterval
	
	// Initial poll
	p.poll(ctx, out)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		
		case <-p.refreshCh:
			// Immediate refresh requested
			p.poll(ctx, out)
			ticker.Reset(p.fastInterval)
			interval = p.fastInterval
		
		case <-ticker.C:
			allRuns := p.poll(ctx, out)
			
			// Adjust polling interval based on activity
			newInterval := p.nextInterval(allRuns, interval)
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
		}
	}
}

// poll fetches builds from all configured projects and emits events
func (p *Provider) poll(ctx context.Context, out chan<- core.RunEvent) []core.Run {
	allRuns := []core.Run{}
	
	for _, project := range p.cfg.Projects {
		select {
		case <-ctx.Done():
			return allRuns
		default:
		}
		
		runs, err := p.client.ListBuilds(p.cfg.Org, project.Name, project.Pipelines)
		if err != nil {
			p.lastError = err
			log.Printf("azure provider error for %s/%s: %v", p.cfg.Org, project.Name, err)
			p.emitEvent(ctx, out, nil, err)
			continue
		}
		
		allRuns = append(allRuns, runs...)
	}
	
	// Emit combined results
	if len(allRuns) > 0 {
		p.lastError = nil
		p.emitEvent(ctx, out, allRuns, nil)
	}
	
	return allRuns
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
