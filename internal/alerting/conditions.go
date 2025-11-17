package alerting

import (
	"context"
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/storage"
)

// WorkflowFailedCondition alerts when a workflow fails
type WorkflowFailedCondition struct{}

func (c *WorkflowFailedCondition) Name() string {
	return "Workflow Failed"
}

func (c *WorkflowFailedCondition) Check(run core.Run, _ *storage.Storage) (bool, string) {
	// Alert on failure or cancellation
	if run.Status == core.RunStatusCompleted && 
	   (run.Conclusion == "failure" || run.Conclusion == "cancelled") {
		return true, fmt.Sprintf("Workflow '%s' %s in %s/%s", 
			run.WorkflowName, run.Conclusion, run.Repo, run.Branch)
	}
	return false, ""
}

// DurationSpikeCondition alerts when build takes significantly longer than average
type DurationSpikeCondition struct {
	Threshold float64 // Multiplier (e.g., 2.0 = 200% of average)
}

func (c *DurationSpikeCondition) Name() string {
	return "Duration Spike"
}

func (c *DurationSpikeCondition) Check(run core.Run, store *storage.Storage) (bool, string) {
	// Only check completed runs
	if run.Status != core.RunStatusCompleted || run.Duration == 0 {
		return false, ""
	}

	// Get historical average for this workflow
	since := time.Now().Add(-7 * 24 * time.Hour) // Last 7 days
	
	history, err := store.GetRunHistory(context.Background(), run.Provider, run.Repo, run.WorkflowName, since, 100)
	if err != nil || len(history) < 3 {
		// Need at least 3 historical runs for meaningful average
		return false, ""
	}

	// Calculate average duration
	var totalDuration time.Duration
	count := 0
	for _, h := range history {
		if h.Status == core.RunStatusCompleted && h.Duration > 0 {
			totalDuration += h.Duration
			count++
		}
	}

	if count < 3 {
		return false, ""
	}

	avgDuration := totalDuration / time.Duration(count)
	threshold := time.Duration(float64(avgDuration) * c.Threshold)

	if run.Duration > threshold {
		return true, fmt.Sprintf("Duration spike: %s (avg: %s, %d%% increase) for '%s' in %s",
			run.Duration.Round(time.Second),
			avgDuration.Round(time.Second),
			int(((float64(run.Duration)-float64(avgDuration))/float64(avgDuration))*100),
			run.WorkflowName,
			run.Repo)
	}

	return false, ""
}

// HighFailureRateCondition alerts when failure rate exceeds threshold
type HighFailureRateCondition struct {
	Threshold float64       // Percentage (e.g., 0.2 = 20%)
	Period    time.Duration // Time window to check (e.g., 24h)
	MinRuns   int           // Minimum runs to evaluate (avoid false positives)
}

func (c *HighFailureRateCondition) Name() string {
	return "High Failure Rate"
}

func (c *HighFailureRateCondition) Check(run core.Run, store *storage.Storage) (bool, string) {
	// Only check on completed runs
	if run.Status != core.RunStatusCompleted {
		return false, ""
	}

	// Get recent runs for this workflow
	since := time.Now().Add(-c.Period)

	history, err := store.GetRunHistory(context.Background(), run.Provider, run.Repo, run.WorkflowName, since, 1000)
	if err != nil || len(history) < c.MinRuns {
		return false, ""
	}

	// Calculate failure rate
	failures := 0
	total := 0
	for _, h := range history {
		if h.Status == core.RunStatusCompleted {
			total++
			if h.Conclusion == "failure" || h.Conclusion == "cancelled" {
				failures++
			}
		}
	}

	if total < c.MinRuns {
		return false, ""
	}

	failureRate := float64(failures) / float64(total)
	if failureRate > c.Threshold {
		return true, fmt.Sprintf("High failure rate: %.0f%% (%d/%d failed) in last %s for '%s' in %s",
			failureRate*100,
			failures,
			total,
			c.Period.String(),
			run.WorkflowName,
			run.Repo)
	}

	return false, ""
}

// BuildQueuedTooLongCondition alerts when builds stay queued for too long
type BuildQueuedTooLongCondition struct {
	MaxQueueTime time.Duration // Max acceptable queue time
}

func (c *BuildQueuedTooLongCondition) Name() string {
	return "Build Queued Too Long"
}

func (c *BuildQueuedTooLongCondition) Check(run core.Run, _ *storage.Storage) (bool, string) {
	// Check if currently queued
	if run.Status != core.RunStatusQueued {
		return false, ""
	}

	queueTime := time.Since(run.StartedAt)
	if queueTime > c.MaxQueueTime {
		return true, fmt.Sprintf("Build queued for %s (max: %s) for '%s' in %s",
			queueTime.Round(time.Second),
			c.MaxQueueTime.String(),
			run.WorkflowName,
			run.Repo)
	}

	return false, ""
}
