package alerting

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/storage"
)

func TestWorkflowFailedCondition(t *testing.T) {
	condition := &WorkflowFailedCondition{}

	tests := []struct {
		name      string
		run       core.Run
		wantAlert bool
	}{
		{
			name: "failed workflow",
			run: core.Run{
				Status:     core.RunStatusCompleted,
				Conclusion: "failure",
				WorkflowName: "CI",
				Repo:       "test/repo",
				Branch:     "main",
			},
			wantAlert: true,
		},
		{
			name: "cancelled workflow",
			run: core.Run{
				Status:     core.RunStatusCompleted,
				Conclusion: "cancelled",
				WorkflowName: "CI",
			},
			wantAlert: true,
		},
		{
			name: "successful workflow",
			run: core.Run{
				Status:     core.RunStatusCompleted,
				Conclusion: "success",
			},
			wantAlert: false,
		},
		{
			name: "in progress workflow",
			run: core.Run{
				Status: core.RunStatusInProgress,
			},
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldAlert, msg := condition.Check(tt.run, nil)
			if shouldAlert != tt.wantAlert {
				t.Errorf("Check() shouldAlert = %v, want %v", shouldAlert, tt.wantAlert)
			}
			if shouldAlert && msg == "" {
				t.Error("Check() returned empty message for alert")
			}
		})
	}
}

func TestDurationSpikeCondition(t *testing.T) {
	// Create temp storage for testing
	store, cleanup := createTestStorage(t)
	defer cleanup()

	// Add historical runs with average duration of 5 minutes
	baseTime := time.Now().Add(-24 * time.Hour)
	for i := 0; i < 5; i++ {
		run := core.Run{
			ID:           string(rune(i)),
			Provider:     "test",
			Repo:         "test/repo",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			Conclusion:   "success",
			Duration:     5 * time.Minute,
			StartedAt:    baseTime.Add(time.Duration(i) * time.Hour),
			UpdatedAt:    baseTime.Add(time.Duration(i)*time.Hour + 5*time.Minute),
		}
		if err := store.Store(context.Background(), run); err != nil {
			t.Fatal(err)
		}
	}

	condition := &DurationSpikeCondition{
		Threshold: 2.0, // 200% of average (10 minutes)
	}

	tests := []struct {
		name      string
		run       core.Run
		wantAlert bool
	}{
		{
			name: "duration spike (11 minutes)",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "CI",
				Status:       core.RunStatusCompleted,
				Duration:     11 * time.Minute,
			},
			wantAlert: true,
		},
		{
			name: "normal duration (5 minutes)",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "CI",
				Status:       core.RunStatusCompleted,
				Duration:     5 * time.Minute,
			},
			wantAlert: false,
		},
		{
			name: "in progress (no duration)",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "CI",
				Status:       core.RunStatusInProgress,
			},
			wantAlert: false,
		},
		{
			name: "different workflow (no history)",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "Deploy",
				Status:       core.RunStatusCompleted,
				Duration:     20 * time.Minute,
			},
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldAlert, msg := condition.Check(tt.run, store)
			if shouldAlert != tt.wantAlert {
				t.Errorf("Check() shouldAlert = %v, want %v (msg: %s)", shouldAlert, tt.wantAlert, msg)
			}
		})
	}
}

func TestHighFailureRateCondition(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	// Add 10 runs: 3 failures, 7 successes = 30% failure rate
	baseTime := time.Now().Add(-24 * time.Hour)
	for i := 0; i < 10; i++ {
		conclusion := "success"
		if i < 3 {
			conclusion = "failure"
		}
		run := core.Run{
			ID:           string(rune(i)),
			Provider:     "test",
			Repo:         "test/repo",
			WorkflowName: "CI",
			Status:       core.RunStatusCompleted,
			Conclusion:   conclusion,
			Duration:     5 * time.Minute,
			StartedAt:    baseTime.Add(time.Duration(i) * time.Hour),
			UpdatedAt:    baseTime.Add(time.Duration(i)*time.Hour + 5*time.Minute),
		}
		if err := store.Store(context.Background(), run); err != nil {
			t.Fatal(err)
		}
	}

	condition := &HighFailureRateCondition{
		Threshold: 0.2,           // 20%
		Period:    48 * time.Hour, // Last 48 hours
		MinRuns:   5,
	}

	tests := []struct {
		name      string
		run       core.Run
		wantAlert bool
	}{
		{
			name: "high failure rate (30% > 20%)",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "CI",
				Status:       core.RunStatusCompleted,
				Conclusion:   "failure",
			},
			wantAlert: true,
		},
		{
			name: "in progress",
			run: core.Run{
				Repo:         "test/repo",
				WorkflowName: "CI",
				Status:       core.RunStatusInProgress,
			},
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldAlert, msg := condition.Check(tt.run, store)
			if shouldAlert != tt.wantAlert {
				t.Errorf("Check() shouldAlert = %v, want %v (msg: %s)", shouldAlert, tt.wantAlert, msg)
			}
		})
	}
}

func TestBuildQueuedTooLongCondition(t *testing.T) {
	condition := &BuildQueuedTooLongCondition{
		MaxQueueTime: 10 * time.Minute,
	}

	tests := []struct {
		name      string
		run       core.Run
		wantAlert bool
	}{
		{
			name: "queued too long (15 minutes)",
			run: core.Run{
				Status:       core.RunStatusQueued,
				StartedAt:    time.Now().Add(-15 * time.Minute),
				WorkflowName: "CI",
				Repo:         "test/repo",
			},
			wantAlert: true,
		},
		{
			name: "queued acceptable time (5 minutes)",
			run: core.Run{
				Status:    core.RunStatusQueued,
				StartedAt: time.Now().Add(-5 * time.Minute),
			},
			wantAlert: false,
		},
		{
			name: "not queued (in progress)",
			run: core.Run{
				Status:    core.RunStatusInProgress,
				StartedAt: time.Now().Add(-30 * time.Minute),
			},
			wantAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldAlert, msg := condition.Check(tt.run, nil)
			if shouldAlert != tt.wantAlert {
				t.Errorf("Check() shouldAlert = %v, want %v (msg: %s)", shouldAlert, tt.wantAlert, msg)
			}
		})
	}
}

// Helper to create test storage
func createTestStorage(t *testing.T) (*storage.Storage, func()) {
	t.Helper()
	store, err := storage.New(storage.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}
	return store, func() {
		store.Close()
	}
}
