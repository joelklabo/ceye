package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestStorageNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		Path:          dbPath,
		RetentionDays: 30,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestStorageStoreAndRetrieve(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	run := core.Run{
		ID:           "test-1",
		Provider:     "github",
		Repo:         "owner/repo",
		WorkflowName: "CI",
		Status:       core.RunStatusCompleted,
		Conclusion:   "success",
		Branch:       "main",
		CommitSHA:    "abc123",
		StartedAt:    time.Now().Add(-5 * time.Minute),
		UpdatedAt:    time.Now(),
		Duration:     5 * time.Minute,
		URL:          "https://example.com/run/1",
	}

	// Store the run
	if err := storage.Store(ctx, run); err != nil {
		t.Fatalf("failed to store run: %v", err)
	}

	// Retrieve it
	history, err := storage.GetRunHistory(ctx, "github", "", "", time.Time{}, 10)
	if err != nil {
		t.Fatalf("failed to get run history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 run, got %d", len(history))
	}

	retrieved := history[0].Run
	if retrieved.ID != run.ID {
		t.Errorf("expected ID %s, got %s", run.ID, retrieved.ID)
	}
	if retrieved.Provider != run.Provider {
		t.Errorf("expected provider %s, got %s", run.Provider, retrieved.Provider)
	}
	if retrieved.Repo != run.Repo {
		t.Errorf("expected repo %s, got %s", run.Repo, retrieved.Repo)
	}
	if retrieved.Status != run.Status {
		t.Errorf("expected status %s, got %s", run.Status, retrieved.Status)
	}
}

func TestStorageBatch(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	runs := []core.Run{
		{
			ID:           "batch-1",
			Provider:     "github",
			Repo:         "owner/repo1",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-10 * time.Minute),
			UpdatedAt:    time.Now(),
			Duration:     5 * time.Minute,
		},
		{
			ID:           "batch-2",
			Provider:     "github",
			Repo:         "owner/repo2",
			WorkflowName: "Deploy",
			Status:       core.RunStatusFailed,
			StartedAt:    time.Now().Add(-8 * time.Minute),
			UpdatedAt:    time.Now(),
			Duration:     3 * time.Minute,
		},
		{
			ID:           "batch-3",
			Provider:     "azure",
			Repo:         "project/repo",
			WorkflowName: "Build",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-6 * time.Minute),
			UpdatedAt:    time.Now(),
			Duration:     4 * time.Minute,
		},
	}

	// Store batch
	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Retrieve all
	history, err := storage.GetRunHistory(ctx, "", "", "", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get run history: %v", err)
	}

	if len(history) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(history))
	}
}

func TestStorageFiltering(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	runs := []core.Run{
		{
			ID:           "filter-1",
			Provider:     "github",
			Repo:         "owner/repo1",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-10 * time.Minute),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "filter-2",
			Provider:     "github",
			Repo:         "owner/repo2",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-8 * time.Minute),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "filter-3",
			Provider:     "azure",
			Repo:         "owner/repo1",
			WorkflowName: "Deploy",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-6 * time.Minute),
			UpdatedAt:    time.Now(),
		},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Filter by provider
	history, err := storage.GetRunHistory(ctx, "github", "", "", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get github runs: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 github runs, got %d", len(history))
	}

	// Filter by repo
	history, err = storage.GetRunHistory(ctx, "", "owner/repo1", "", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get repo1 runs: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 repo1 runs, got %d", len(history))
	}

	// Filter by workflow
	history, err = storage.GetRunHistory(ctx, "", "", "CI", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get CI runs: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 CI runs, got %d", len(history))
	}

	// Combined filters
	history, err = storage.GetRunHistory(ctx, "github", "owner/repo1", "CI", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get filtered runs: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 filtered run, got %d", len(history))
	}
	if history[0].ID != "filter-1" {
		t.Errorf("expected filter-1, got %s", history[0].ID)
	}
}

func TestStorageTimeFiltering(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	
	runs := []core.Run{
		{
			ID:        "time-1",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCompleted,
			StartedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        "time-2",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCompleted,
			StartedAt: now.Add(-30 * time.Minute),
			UpdatedAt: now.Add(-30 * time.Minute),
		},
		{
			ID:        "time-3",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCompleted,
			StartedAt: now.Add(-5 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
		},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Get runs from last hour
	since := now.Add(-1 * time.Hour)
	history, err := storage.GetRunHistory(ctx, "", "", "", since, 100)
	if err != nil {
		t.Fatalf("failed to get recent runs: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 runs from last hour, got %d", len(history))
	}
}

func TestStorageLimit(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	
	// Store 10 runs
	runs := make([]core.Run, 10)
	for i := 0; i < 10; i++ {
		runs[i] = core.Run{
			ID:           string(rune('a' + i)),
			Provider:     "github",
			Repo:         "owner/repo",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			StartedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
			UpdatedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Limit to 5
	history, err := storage.GetRunHistory(ctx, "", "", "", time.Time{}, 5)
	if err != nil {
		t.Fatalf("failed to get limited runs: %v", err)
	}
	if len(history) != 5 {
		t.Errorf("expected 5 runs, got %d", len(history))
	}
}

func TestStorageMetrics(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	
	runs := []core.Run{
		{
			ID:        "metrics-1",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCompleted,
			StartedAt: now.Add(-10 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
			Duration:  5 * time.Minute,
		},
		{
			ID:        "metrics-2",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusFailed,
			StartedAt: now.Add(-8 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
			Duration:  3 * time.Minute,
		},
		{
			ID:        "metrics-3",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCompleted,
			StartedAt: now.Add(-6 * time.Minute),
			UpdatedAt: now.Add(-2 * time.Minute),
			Duration:  4 * time.Minute,
		},
		{
			ID:        "metrics-4",
			Provider:  "github",
			Repo:      "owner/repo",
			WorkflowName: "CI",
			Status:    core.RunStatusCancelled,
			StartedAt: now.Add(-4 * time.Minute),
			UpdatedAt: now.Add(-3 * time.Minute),
			Duration:  1 * time.Minute,
		},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Get metrics
	metrics, err := storage.GetProviderMetrics(ctx, "github", now.Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	if metrics.Total != 4 {
		t.Errorf("expected 4 total runs, got %d", metrics.Total)
	}
	if metrics.Completed != 2 {
		t.Errorf("expected 2 completed runs, got %d", metrics.Completed)
	}
	if metrics.Failed != 1 {
		t.Errorf("expected 1 failed run, got %d", metrics.Failed)
	}
	if metrics.Cancelled != 1 {
		t.Errorf("expected 1 cancelled run, got %d", metrics.Cancelled)
	}

	expectedSuccessRate := 0.5 // 2 completed out of 4 total
	if metrics.SuccessRate != expectedSuccessRate {
		t.Errorf("expected success rate %.2f, got %.2f", expectedSuccessRate, metrics.SuccessRate)
	}

	// Average duration should be (5 + 3 + 4 + 1) / 4 = 3.25 minutes
	expectedAvg := (5 + 3 + 4 + 1) * time.Minute / 4
	tolerance := 1 * time.Second
	if metrics.AvgDuration < expectedAvg-tolerance || metrics.AvgDuration > expectedAvg+tolerance {
		t.Errorf("expected avg duration ~%v, got %v", expectedAvg, metrics.AvgDuration)
	}
}

func TestStorageFailureRate(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	
	runs := []core.Run{
		{ID: "fail-1", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-10 * time.Minute), UpdatedAt: now},
		{ID: "fail-2", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-8 * time.Minute), UpdatedAt: now},
		{ID: "fail-3", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-6 * time.Minute), UpdatedAt: now},
		{ID: "fail-4", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-4 * time.Minute), UpdatedAt: now},
		{ID: "fail-5", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-2 * time.Minute), UpdatedAt: now},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Calculate failure rate
	rate, err := storage.GetFailureRate(ctx, "owner/repo", "CI", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to get failure rate: %v", err)
	}

	expectedRate := 0.4 // 2 failures out of 5 total
	if rate != expectedRate {
		t.Errorf("expected failure rate %.2f, got %.2f", expectedRate, rate)
	}
}

func TestStorageAverageDuration(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	
	runs := []core.Run{
		{ID: "dur-1", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-10 * time.Minute), UpdatedAt: now, Duration: 5 * time.Minute},
		{ID: "dur-2", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-8 * time.Minute), UpdatedAt: now, Duration: 3 * time.Minute},
		{ID: "dur-3", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-6 * time.Minute), UpdatedAt: now, Duration: 4 * time.Minute},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Calculate average duration
	avg, err := storage.GetAverageDuration(ctx, "owner/repo", "CI", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to get average duration: %v", err)
	}

	expectedAvg := 4 * time.Minute // (5 + 3 + 4) / 3
	tolerance := 1 * time.Second
	if avg < expectedAvg-tolerance || avg > expectedAvg+tolerance {
		t.Errorf("expected avg duration ~%v, got %v", expectedAvg, avg)
	}
}

func TestStorageCleanup(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()
	
	runs := []core.Run{
		{ID: "old-1", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.AddDate(0, 0, -40), UpdatedAt: now.AddDate(0, 0, -40)},
		{ID: "old-2", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.AddDate(0, 0, -35), UpdatedAt: now.AddDate(0, 0, -35)},
		{ID: "new-1", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
		{ID: "new-2", Provider: "github", Repo: "owner/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-30 * time.Minute), UpdatedAt: now},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store batch: %v", err)
	}

	// Cleanup runs older than 30 days
	deleted, err := storage.Cleanup(ctx, 30)
	if err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deleted runs, got %d", deleted)
	}

	// Verify remaining runs
	history, err := storage.GetRunHistory(ctx, "", "", "", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get run history: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("expected 2 remaining runs, got %d", len(history))
	}
}

func TestStorageUpdateRun(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	ctx := context.Background()
	
	// Store initial run
	run := core.Run{
		ID:           "update-1",
		Provider:     "github",
		Repo:         "owner/repo",
		WorkflowName: "CI",
		Status:       core.RunStatusInProgress,
		StartedAt:    time.Now().Add(-5 * time.Minute),
		UpdatedAt:    time.Now(),
	}

	if err := storage.Store(ctx, run); err != nil {
		t.Fatalf("failed to store run: %v", err)
	}

	// Update the run (change status)
	run.Status = core.RunStatusCompleted
	run.Conclusion = "success"
	run.Duration = 5 * time.Minute
	run.UpdatedAt = time.Now()

	if err := storage.Store(ctx, run); err != nil {
		t.Fatalf("failed to update run: %v", err)
	}

	// Retrieve and verify
	history, err := storage.GetRunHistory(ctx, "", "", "", time.Time{}, 100)
	if err != nil {
		t.Fatalf("failed to get run history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 run, got %d (should be updated, not duplicated)", len(history))
	}

	retrieved := history[0].Run
	if retrieved.Status != core.RunStatusCompleted {
		t.Errorf("expected status completed, got %s", retrieved.Status)
	}
	if retrieved.Conclusion != "success" {
		t.Errorf("expected conclusion success, got %s", retrieved.Conclusion)
	}
}

// Helper function to create a test storage instance
func newTestStorage(t *testing.T) *Storage {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		Path: dbPath,
	}

	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}

	return storage
}
