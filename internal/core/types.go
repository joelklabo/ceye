package core

import (
	"context"
	"time"
)

// RunStatus represents the lifecycle state of a CI workflow/build run.
type RunStatus string

const (
	RunStatusUnknown    RunStatus = "unknown"
	RunStatusQueued     RunStatus = "queued"
	RunStatusInProgress RunStatus = "in_progress"
	RunStatusCompleted  RunStatus = "completed"
	RunStatusFailed     RunStatus = "failed"
	RunStatusCancelled  RunStatus = "cancelled"
)

// Run defines a normalized CI run regardless of provider specifics.
type Run struct {
	ID           string
	Provider     string
	Repo         string
	WorkflowName string
	Status       RunStatus
	Conclusion   string
	Branch       string
	CommitSHA    string
	StartedAt    time.Time
	UpdatedAt    time.Time
	Duration     time.Duration
	URL          string
}

// RunEvent batches run updates emitted by providers.
type RunEvent struct {
	Provider  string
	Runs      []Run
	Timestamp time.Time
	Err       error
	Message   string
	Health    map[string]ProviderHealth
}

// Provider is implemented by CI backends (GitHub, Azure, etc.).
type Provider interface {
	Name() string
	Start(ctx context.Context, out chan<- RunEvent) error
}

// ProviderHealth tracks recent health details for a provider.
type ProviderHealth struct {
	LastError   time.Time
	ErrorCount  int
	LastSuccess time.Time
}
