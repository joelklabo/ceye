package core

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStoreMergeAddsRuns(t *testing.T) {
	store := NewStore()
	now := time.Now()

	event := RunEvent{
		Provider: "github",
		Runs: []Run{
			{
				ID:           "1001",
				Provider:     "github",
				Repo:         "example/repo",
				WorkflowName: "CI",
				Status:       RunStatusInProgress,
				Branch:       "main",
				CommitSHA:    "abc123",
				StartedAt:    now.Add(-1 * time.Minute),
				UpdatedAt:    now,
				URL:          "https://example.com/runs/1001",
			},
		},
		Timestamp: now,
	}

	store.Merge(event)

	runs := store.ListRuns("")
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.ID != "1001" || got.Status != RunStatusInProgress {
		t.Fatalf("unexpected run contents: %+v", got)
	}
}

func TestStoreMergeUpdatesExistingRuns(t *testing.T) {
	store := NewStore()
	base := time.Now().Add(-2 * time.Minute)

	first := Run{
		ID:           "2001",
		Provider:     "github",
		Repo:         "example/repo",
		WorkflowName: "Deploy",
		Status:       RunStatusInProgress,
		StartedAt:    base,
		UpdatedAt:    base,
	}

	store.Merge(RunEvent{Provider: "github", Runs: []Run{first}, Timestamp: base})

	updated := first
	updated.Status = RunStatusCompleted
	updated.Conclusion = "success"
	updated.UpdatedAt = base.Add(90 * time.Second)

	store.Merge(RunEvent{Provider: "github", Runs: []Run{updated}, Timestamp: updated.UpdatedAt})

	runs := store.ListRuns("github")
	if len(runs) != 1 {
		t.Fatalf("expected 1 run after update, got %d", len(runs))
	}

	got := runs[0]
	if got.Status != RunStatusCompleted {
		t.Fatalf("expected completed status, got %s", got.Status)
	}
	if got.Conclusion != "success" {
		t.Fatalf("expected updated conclusion, got %s", got.Conclusion)
	}
	if got.UpdatedAt != updated.UpdatedAt {
		t.Fatalf("expected updated timestamp %v, got %v", updated.UpdatedAt, got.UpdatedAt)
	}
}

func TestStoreListRunsFiltersAndSorts(t *testing.T) {
	store := NewStore()
	base := time.Now()

	runs := []Run{
		{
			ID:           "g1",
			Provider:     "github",
			WorkflowName: "Build",
			Status:       RunStatusCompleted,
			UpdatedAt:    base.Add(1 * time.Minute),
		},
		{
			ID:           "g2",
			Provider:     "github",
			WorkflowName: "Test",
			Status:       RunStatusInProgress,
			UpdatedAt:    base.Add(2 * time.Minute),
		},
		{
			ID:           "a1",
			Provider:     "azure",
			WorkflowName: "CI",
			Status:       RunStatusQueued,
			UpdatedAt:    base.Add(30 * time.Second),
		},
	}

	store.Merge(RunEvent{Provider: "github", Runs: runs[:2], Timestamp: base})
	store.Merge(RunEvent{Provider: "azure", Runs: []Run{runs[2]}, Timestamp: base})

	githubRuns := store.ListRuns("github")
	if len(githubRuns) != 2 {
		t.Fatalf("expected 2 github runs, got %d", len(githubRuns))
	}

	if githubRuns[0].ID != "g2" {
		t.Fatalf("expected most recent github run first, got %s", githubRuns[0].ID)
	}
	if githubRuns[1].ID != "g1" {
		t.Fatalf("unexpected order for github runs: %s then %s", githubRuns[0].ID, githubRuns[1].ID)
	}

	allRuns := store.ListRuns("")
	if len(allRuns) != 3 {
		t.Fatalf("expected 3 runs when no filter, got %d", len(allRuns))
	}
	if allRuns[0].ID != "g2" || allRuns[1].ID != "a1" || allRuns[2].ID != "g1" {
		t.Fatalf("unexpected order for all runs: %s %s %s", allRuns[0].ID, allRuns[1].ID, allRuns[2].ID)
	}
}

// Mock storage backend for testing
type mockStorage struct {
	mu    sync.Mutex
	calls [][]Run // Each call to StoreBatch appends a slice
}

func (m *mockStorage) Store(ctx context.Context, run Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, []Run{run})
	return nil
}

func (m *mockStorage) StoreBatch(ctx context.Context, runs []Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, runs)
	return nil
}

func (m *mockStorage) getCalls() [][]Run {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestStoreWithStoragePersistsCompletedRuns(t *testing.T) {
	mock := &mockStorage{}
	store := NewStoreWithStorage(mock)
	now := time.Now()

	// Merge an in-progress run (should not be stored)
	inProgressRun := Run{
		ID:           "test-1",
		Provider:     "github",
		Repo:         "owner/repo",
		WorkflowName: "CI",
		Status:       RunStatusInProgress,
		StartedAt:    now.Add(-2 * time.Minute),
		UpdatedAt:    now,
	}

	store.Merge(RunEvent{
		Provider:  "github",
		Runs:      []Run{inProgressRun},
		Timestamp: now,
	})

	time.Sleep(50 * time.Millisecond) // Give goroutine time to run

	calls := mock.getCalls()
	if len(calls) != 0 {
		t.Errorf("expected 0 storage calls for in-progress run, got %d", len(calls))
	}

	// Now complete the run (should be stored)
	completedRun := inProgressRun
	completedRun.Status = RunStatusCompleted
	completedRun.Conclusion = "success"
	completedRun.Duration = 2 * time.Minute
	completedRun.UpdatedAt = now.Add(10 * time.Second)

	store.Merge(RunEvent{
		Provider:  "github",
		Runs:      []Run{completedRun},
		Timestamp: completedRun.UpdatedAt,
	})

	time.Sleep(50 * time.Millisecond) // Give goroutine time to run

	calls = mock.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 storage call for completed run, got %d", len(calls))
	}

	stored := calls[0]
	if len(stored) != 1 {
		t.Fatalf("expected 1 run in storage call, got %d", len(stored))
	}

	if stored[0].ID != "test-1" {
		t.Errorf("expected stored run ID test-1, got %s", stored[0].ID)
	}
	if stored[0].Status != RunStatusCompleted {
		t.Errorf("expected stored run status completed, got %s", stored[0].Status)
	}
}

func TestStoreWithStorageDoesNotDuplicateCompletedRuns(t *testing.T) {
	mock := &mockStorage{}
	store := NewStoreWithStorage(mock)
	now := time.Now()

	completedRun := Run{
		ID:           "test-2",
		Provider:     "github",
		Repo:         "owner/repo",
		WorkflowName: "CI",
		Status:       RunStatusCompleted,
		Conclusion:   "success",
		StartedAt:    now.Add(-5 * time.Minute),
		UpdatedAt:    now,
		Duration:     5 * time.Minute,
	}

	// Merge the completed run twice
	store.Merge(RunEvent{
		Provider:  "github",
		Runs:      []Run{completedRun},
		Timestamp: now,
	})

	time.Sleep(50 * time.Millisecond)

	calls := mock.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 storage call for first completion, got %d", len(calls))
	}

	// Merge again (should not store again)
	store.Merge(RunEvent{
		Provider:  "github",
		Runs:      []Run{completedRun},
		Timestamp: now.Add(1 * time.Second),
	})

	time.Sleep(50 * time.Millisecond)

	calls = mock.getCalls()
	if len(calls) != 1 {
		t.Errorf("expected still 1 storage call (no duplicate), got %d", len(calls))
	}
}

func TestStoreWithStorageBatchesSameEvent(t *testing.T) {
	mock := &mockStorage{}
	store := NewStoreWithStorage(mock)
	now := time.Now()

	runs := []Run{
		{
			ID:           "batch-1",
			Provider:     "github",
			Repo:         "owner/repo",
			WorkflowName: "CI",
			Status:       RunStatusCompleted,
			StartedAt:    now.Add(-5 * time.Minute),
			UpdatedAt:    now,
		},
		{
			ID:           "batch-2",
			Provider:     "github",
			Repo:         "owner/repo",
			WorkflowName: "Deploy",
			Status:       RunStatusCompleted,
			StartedAt:    now.Add(-3 * time.Minute),
			UpdatedAt:    now,
		},
	}

	store.Merge(RunEvent{
		Provider:  "github",
		Runs:      runs,
		Timestamp: now,
	})

	time.Sleep(50 * time.Millisecond)

	calls := mock.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 storage call (batched), got %d", len(calls))
	}

	stored := calls[0]
	if len(stored) != 2 {
		t.Errorf("expected 2 runs in batch, got %d", len(stored))
	}
}
