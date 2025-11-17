package validation

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers/github"
)

// TestHarnessBasicSetup tests that we can create and start a validation harness
func TestHarnessBasicSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	// Create a harness with mock repos
	repos := []github.RepoConfig{
		{Owner: "test", Repo: "repo1"},
	}

	harness, err := NewHarness(repos)
	if err != nil {
		t.Fatalf("Failed to create harness: %v", err)
	}

	// Start harness for a short duration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = harness.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Harness Start() failed: %v", err)
	}

	// Verify both stores were created
	if harness.WebhookStore() == nil {
		t.Error("Webhook store not initialized")
	}
	if harness.PollingStore() == nil {
		t.Error("Polling store not initialized")
	}
}

// TestHarnessComparison tests that the harness can compare stores
func TestHarnessComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	repos := []github.RepoConfig{
		{Owner: "test", Repo: "repo1"},
	}

	harness, err := NewHarness(repos)
	if err != nil {
		t.Fatalf("Failed to create harness: %v", err)
	}

	// Manually add test data to both stores
	testRun := core.Run{
		ID:           "test123",
		Provider:     "github",
		Repo:         "test/repo1",
		WorkflowName: "CI",
		Status:       core.RunStatusCompleted,
		Conclusion:   "success",
	}

	// Add to both stores - should match
	harness.WebhookStore().Merge(core.RunEvent{
		Provider:  "github",
		Runs:      []core.Run{testRun},
		Timestamp: time.Now(),
	})

	harness.PollingStore().Merge(core.RunEvent{
		Provider:  "github",
		Runs:      []core.Run{testRun},
		Timestamp: time.Now(),
	})

	// Compare stores
	discrepancies := harness.Compare()
	if len(discrepancies) > 0 {
		t.Errorf("Expected 0 discrepancies with matching data, got %d", len(discrepancies))
	}

	// Test with missing run in webhook store
	testRun2 := testRun
	testRun2.ID = "test456"
	harness.PollingStore().Merge(core.RunEvent{
		Provider:  "github",
		Runs:      []core.Run{testRun2},
		Timestamp: time.Now(),
	})

	discrepancies = harness.Compare()
	if len(discrepancies) == 0 {
		t.Error("Expected discrepancies when polling store has extra run")
	}

	// Verify discrepancy type
	found := false
	for _, d := range discrepancies {
		if d.Type == "missing_in_webhook" && d.RunID == "test456" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'missing_in_webhook' discrepancy for test456")
	}
}

// TestHarnessMetrics tests that metrics are collected correctly
func TestHarnessMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	repos := []github.RepoConfig{
		{Owner: "test", Repo: "repo1"},
	}

	harness, err := NewHarness(repos)
	if err != nil {
		t.Fatalf("Failed to create harness: %v", err)
	}

	// Add test data with timestamps
	now := time.Now()
	testRun := core.Run{
		ID:           "test123",
		Provider:     "github",
		Repo:         "test/repo1",
		WorkflowName: "CI",
		Status:       core.RunStatusCompleted,
		Conclusion:   "success",
		UpdatedAt:    now,
	}

	// Webhook arrives "now" - add to store
	harness.WebhookStore().Merge(core.RunEvent{
		Provider:  "github",
		Runs:      []core.Run{testRun},
		Timestamp: now,
	})
	harness.RecordWebhookEvent(testRun, now)

	// Polling discovers it 30 seconds later - add to store
	harness.PollingStore().Merge(core.RunEvent{
		Provider:  "github",
		Runs:      []core.Run{testRun},
		Timestamp: now.Add(30 * time.Second),
	})
	harness.RecordPollingEvent(testRun, now.Add(30*time.Second))

	metrics := harness.GetMetrics()

	if metrics.WebhookRunCount != 1 {
		t.Errorf("Expected 1 webhook run, got %d", metrics.WebhookRunCount)
	}

	if metrics.PollingRunCount != 1 {
		t.Errorf("Expected 1 polling run, got %d", metrics.PollingRunCount)
	}

	if metrics.MatchedRuns != 1 {
		t.Errorf("Expected 1 matched run, got %d", metrics.MatchedRuns)
	}

	// Check that webhook was faster
	if metrics.WebhookAdvantage < 25*time.Second {
		t.Errorf("Expected webhook advantage ~30s, got %v", metrics.WebhookAdvantage)
	}
}
