package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

const (
	defaultRunCount = 3
	defaultInterval = 5 * time.Second
)

type Provider struct {
	name     string
	interval time.Duration
	runs     []demoRun
	tick     int
}

type demoRun struct {
	repo     string
	workflow string
	branch   string
	commit   string
	url      string
	id       string
}

// New creates a demo provider that synthesizes run updates.
func New(count int, interval time.Duration) core.Provider {
	if count <= 0 {
		count = defaultRunCount
	}
	if interval <= 0 {
		interval = defaultInterval
	}
	runs := make([]demoRun, count)
	for i := 0; i < count; i++ {
		runs[i] = demoRun{
			repo:     fmt.Sprintf("example/service-%d", i%3+1),
			workflow: []string{"Build", "Test", "Deploy"}[i%3],
			branch:   []string{"main", "develop", "release"}[i%3],
			commit:   fmt.Sprintf("d34db33f%02d", i),
			url:      fmt.Sprintf("https://example.com/demo/%d", i+1),
			id:       fmt.Sprintf("demo-%d", i+1),
		}
	}
	return &Provider{
		name:     "demo",
		interval: interval,
		runs:     runs,
	}
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	// emit initial event immediately
	p.emit(ctx, out, time.Now())

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
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
	runs := make([]core.Run, len(p.runs))
	for i, sample := range p.runs {
		run := core.Run{
			ID:           sample.id,
			Provider:     p.name,
			Repo:         sample.repo,
			WorkflowName: sample.workflow,
			Branch:       sample.branch,
			CommitSHA:    sample.commit,
			URL:          sample.url,
		}
		p.applyState(&run, i, ts)
		runs[i] = run
	}

	event := core.RunEvent{Provider: p.name, Runs: runs, Timestamp: ts}
	select {
	case <-ctx.Done():
		return
	case out <- event:
		return
	}
}

func (p *Provider) applyState(run *core.Run, idx int, ts time.Time) {
	state := (p.tick + idx) % 4
	switch state {
	case 0:
		run.Status = core.RunStatusQueued
		run.StartedAt = ts
	case 1:
		run.Status = core.RunStatusInProgress
		run.StartedAt = ts.Add(-time.Duration(idx+1) * time.Minute)
	case 2:
		run.Status = core.RunStatusCompleted
		run.Conclusion = "success"
		run.StartedAt = ts.Add(-5 * time.Minute)
		run.Duration = 5 * time.Minute
	default:
		run.Status = core.RunStatusFailed
		run.Conclusion = "failed"
		run.StartedAt = ts.Add(-4 * time.Minute)
		run.Duration = 4 * time.Minute
	}
	run.UpdatedAt = ts
}
