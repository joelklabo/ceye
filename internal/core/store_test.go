package core

import (
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
