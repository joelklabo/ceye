package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// Provider models a synthetic GitLab pipeline provider.
type Provider struct {
	name     string
	url      string
	interval time.Duration
	tick     int
}

// Config defines GitLab pipeline settings.
type Config struct {
	Project string `mapstructure:"project"`
}

// NewProvider constructs a GitLab provider stub.
func NewProvider(cfg Config) core.Provider {
	name := "gitlab"
	if cfg.Project != "" {
		name = fmt.Sprintf("gitlab:%s", cfg.Project)
	}
	return &Provider{name: name, url: fmt.Sprintf("https://gitlab.com/%s/-/pipelines/1", cfg.Project), interval: 7 * time.Second}
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	p.emit(ctx, out, time.Now())
	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			p.emit(ctx, out, t)
		}
	}
}

func (p *Provider) emit(ctx context.Context, out chan<- core.RunEvent, ts time.Time) {
	p.tick++
	run := core.Run{
		ID:           fmt.Sprintf("gitlab-run-%d", p.tick),
		Provider:     "gitlab",
		Repo:         p.name,
		WorkflowName: "pipeline",
		Branch:       "main",
		CommitSHA:    fmt.Sprintf("glsha%d", p.tick),
		URL:          p.url,
		Status:       core.RunStatusQueued,
		UpdatedAt:    ts,
	}
	if p.tick%3 == 0 {
		run.Status = core.RunStatusCompleted
		run.Conclusion = "success"
		run.Duration = 90 * time.Second
		run.UpdatedAt = ts.Add(-5 * time.Second)
	} else if p.tick%3 == 1 {
		run.Status = core.RunStatusInProgress
	} else {
		run.Status = core.RunStatusFailed
		run.Conclusion = "failed"
	}
	event := core.RunEvent{Provider: "gitlab", Runs: []core.Run{run}, Timestamp: ts}
	select {
	case <-ctx.Done():
		return
	case out <- event:
	}
}
