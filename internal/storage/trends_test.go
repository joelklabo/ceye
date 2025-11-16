package storage

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestGetSuccessRateTrend(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Create data for two periods
	// Previous period (14-7 days ago): 80% success (4/5)
	previousStart := now.AddDate(0, 0, -14)
	previousRuns := []core.Run{
		{ID: "p1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart, UpdatedAt: previousStart},
		{ID: "p2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(1 * time.Hour), UpdatedAt: previousStart.Add(1 * time.Hour)},
		{ID: "p3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(2 * time.Hour), UpdatedAt: previousStart.Add(2 * time.Hour)},
		{ID: "p4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(3 * time.Hour), UpdatedAt: previousStart.Add(3 * time.Hour)},
		{ID: "p5", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: previousStart.Add(4 * time.Hour), UpdatedAt: previousStart.Add(4 * time.Hour)},
	}

	// Current period (last 7 days): 100% success (5/5)
	currentStart := now.AddDate(0, 0, -7)
	currentRuns := []core.Run{
		{ID: "c1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart, UpdatedAt: currentStart},
		{ID: "c2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(1 * time.Hour), UpdatedAt: currentStart.Add(1 * time.Hour)},
		{ID: "c3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(2 * time.Hour), UpdatedAt: currentStart.Add(2 * time.Hour)},
		{ID: "c4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(3 * time.Hour), UpdatedAt: currentStart.Add(3 * time.Hour)},
		{ID: "c5", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(4 * time.Hour), UpdatedAt: currentStart.Add(4 * time.Hour)},
	}

	if err := storage.StoreBatch(ctx, previousRuns); err != nil {
		t.Fatalf("failed to store previous runs: %v", err)
	}
	if err := storage.StoreBatch(ctx, currentRuns); err != nil {
		t.Fatalf("failed to store current runs: %v", err)
	}

	// Get trend for 7-day period
	trend, err := analyzer.GetSuccessRateTrend(ctx, "github", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get success rate trend: %v", err)
	}

	if trend.Metric != "success_rate" {
		t.Errorf("expected metric 'success_rate', got %s", trend.Metric)
	}

	// Current should be 1.0 (100%)
	if trend.Current != 1.0 {
		t.Errorf("expected current success rate 1.0, got %.2f", trend.Current)
	}

	// Previous should be 0.8 (80%)
	if trend.Previous != 0.8 {
		t.Errorf("expected previous success rate 0.8, got %.2f", trend.Previous)
	}

	// Change should be positive (improvement)
	if trend.Change <= 0 {
		t.Errorf("expected positive change, got %.2f", trend.Change)
	}

	if trend.Direction != TrendUp {
		t.Errorf("expected trend up, got %v", trend.Direction)
	}
}

func TestGetDurationTrend(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Previous period: average 10 minutes
	previousStart := now.AddDate(0, 0, -14)
	previousRuns := []core.Run{
		{ID: "p1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart, UpdatedAt: previousStart, Duration: 10 * time.Minute},
		{ID: "p2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(1 * time.Hour), UpdatedAt: previousStart.Add(1 * time.Hour), Duration: 10 * time.Minute},
	}

	// Current period: average 5 minutes (50% faster)
	currentStart := now.AddDate(0, 0, -7)
	currentRuns := []core.Run{
		{ID: "c1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart, UpdatedAt: currentStart, Duration: 5 * time.Minute},
		{ID: "c2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(1 * time.Hour), UpdatedAt: currentStart.Add(1 * time.Hour), Duration: 5 * time.Minute},
	}

	if err := storage.StoreBatch(ctx, previousRuns); err != nil {
		t.Fatalf("failed to store previous runs: %v", err)
	}
	if err := storage.StoreBatch(ctx, currentRuns); err != nil {
		t.Fatalf("failed to store current runs: %v", err)
	}

	trend, err := analyzer.GetDurationTrend(ctx, "github", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get duration trend: %v", err)
	}

	if trend.Metric != "duration" {
		t.Errorf("expected metric 'duration', got %s", trend.Metric)
	}

	// Current should be ~300 seconds (5 minutes)
	expectedCurrent := 300.0
	if trend.Current < expectedCurrent-1 || trend.Current > expectedCurrent+1 {
		t.Errorf("expected current duration ~%.0f seconds, got %.2f", expectedCurrent, trend.Current)
	}

	// Previous should be ~600 seconds (10 minutes)
	expectedPrevious := 600.0
	if trend.Previous < expectedPrevious-1 || trend.Previous > expectedPrevious+1 {
		t.Errorf("expected previous duration ~%.0f seconds, got %.2f", expectedPrevious, trend.Previous)
	}

	// Change should be negative (faster is better)
	if trend.Change >= 0 {
		t.Errorf("expected negative change (improvement), got %.2f", trend.Change)
	}

	// Direction should be up (improvement)
	if trend.Direction != TrendUp {
		t.Errorf("expected trend up (improvement), got %v", trend.Direction)
	}
}

func TestGetFrequencyTrend(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Previous period: 2 runs
	previousStart := now.AddDate(0, 0, -14)
	previousRuns := []core.Run{
		{ID: "p1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart, UpdatedAt: previousStart},
		{ID: "p2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(1 * time.Hour), UpdatedAt: previousStart.Add(1 * time.Hour)},
	}

	// Current period: 5 runs (2.5x increase)
	currentStart := now.AddDate(0, 0, -7)
	currentRuns := []core.Run{
		{ID: "c1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart, UpdatedAt: currentStart},
		{ID: "c2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(1 * time.Hour), UpdatedAt: currentStart.Add(1 * time.Hour)},
		{ID: "c3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(2 * time.Hour), UpdatedAt: currentStart.Add(2 * time.Hour)},
		{ID: "c4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(3 * time.Hour), UpdatedAt: currentStart.Add(3 * time.Hour)},
		{ID: "c5", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(4 * time.Hour), UpdatedAt: currentStart.Add(4 * time.Hour)},
	}

	if err := storage.StoreBatch(ctx, previousRuns); err != nil {
		t.Fatalf("failed to store previous runs: %v", err)
	}
	if err := storage.StoreBatch(ctx, currentRuns); err != nil {
		t.Fatalf("failed to store current runs: %v", err)
	}

	trend, err := analyzer.GetFrequencyTrend(ctx, "github", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get frequency trend: %v", err)
	}

	if trend.Metric != "frequency" {
		t.Errorf("expected metric 'frequency', got %s", trend.Metric)
	}

	// Current should be higher than previous
	if trend.Current <= trend.Previous {
		t.Errorf("expected current frequency %.2f > previous %.2f", trend.Current, trend.Previous)
	}

	// Change should be positive
	if trend.Change <= 0 {
		t.Errorf("expected positive change, got %.2f", trend.Change)
	}

	if trend.Direction != TrendUp {
		t.Errorf("expected trend up, got %v", trend.Direction)
	}
}

func TestGetFailureRateTrend(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Previous period: 50% failure rate (2/4 failed)
	previousStart := now.AddDate(0, 0, -14)
	previousRuns := []core.Run{
		{ID: "p1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart, UpdatedAt: previousStart},
		{ID: "p2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: previousStart.Add(1 * time.Hour), UpdatedAt: previousStart.Add(1 * time.Hour)},
		{ID: "p3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: previousStart.Add(2 * time.Hour), UpdatedAt: previousStart.Add(2 * time.Hour)},
		{ID: "p4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: previousStart.Add(3 * time.Hour), UpdatedAt: previousStart.Add(3 * time.Hour)},
	}

	// Current period: 20% failure rate (1/5 failed) - improvement
	currentStart := now.AddDate(0, 0, -7)
	currentRuns := []core.Run{
		{ID: "c1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart, UpdatedAt: currentStart},
		{ID: "c2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(1 * time.Hour), UpdatedAt: currentStart.Add(1 * time.Hour)},
		{ID: "c3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: currentStart.Add(2 * time.Hour), UpdatedAt: currentStart.Add(2 * time.Hour)},
		{ID: "c4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(3 * time.Hour), UpdatedAt: currentStart.Add(3 * time.Hour)},
		{ID: "c5", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: currentStart.Add(4 * time.Hour), UpdatedAt: currentStart.Add(4 * time.Hour)},
	}

	if err := storage.StoreBatch(ctx, previousRuns); err != nil {
		t.Fatalf("failed to store previous runs: %v", err)
	}
	if err := storage.StoreBatch(ctx, currentRuns); err != nil {
		t.Fatalf("failed to store current runs: %v", err)
	}

	trend, err := analyzer.GetFailureRateTrend(ctx, "github", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get failure rate trend: %v", err)
	}

	if trend.Metric != "failure_rate" {
		t.Errorf("expected metric 'failure_rate', got %s", trend.Metric)
	}

	// Current should be 0.2 (20%)
	expectedCurrent := 0.2
	if trend.Current < expectedCurrent-0.01 || trend.Current > expectedCurrent+0.01 {
		t.Errorf("expected current failure rate ~%.2f, got %.2f", expectedCurrent, trend.Current)
	}

	// Previous should be 0.5 (50%)
	expectedPrevious := 0.5
	if trend.Previous < expectedPrevious-0.01 || trend.Previous > expectedPrevious+0.01 {
		t.Errorf("expected previous failure rate ~%.2f, got %.2f", expectedPrevious, trend.Previous)
	}

	// Change should be negative (fewer failures)
	if trend.Change >= 0 {
		t.Errorf("expected negative change (improvement), got %.2f", trend.Change)
	}

	// Direction should be up (improvement - fewer failures)
	if trend.Direction != TrendUp {
		t.Errorf("expected trend up (improvement), got %v", trend.Direction)
	}
}

func TestGetAllTrends(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Add some sample data
	runs := []core.Run{
		{ID: "1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour), Duration: 5 * time.Minute},
		{ID: "2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), Duration: 6 * time.Minute},
		{ID: "3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour), Duration: 4 * time.Minute},
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store runs: %v", err)
	}

	trends, err := analyzer.GetAllTrends(ctx, "github", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get all trends: %v", err)
	}

	expectedMetrics := []string{"success_rate", "duration", "frequency", "failure_rate"}
	for _, metric := range expectedMetrics {
		if _, ok := trends[metric]; !ok {
			t.Errorf("expected trend for metric %s, but not found", metric)
		}
	}

	if len(trends) != 4 {
		t.Errorf("expected 4 trends, got %d", len(trends))
	}
}

func TestDetectAnomalies(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()
	analyzer := NewTrendAnalyzer(storage)

	ctx := context.Background()
	now := time.Now()

	// Create data with high failure rate (should trigger anomaly)
	runs := []core.Run{
		{ID: "1", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusCompleted, StartedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour), Duration: 5 * time.Minute},
		{ID: "2", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), Duration: 6 * time.Minute},
		{ID: "3", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour), Duration: 4 * time.Minute},
		{ID: "4", Provider: "github", Repo: "test/repo", WorkflowName: "CI", Status: core.RunStatusFailed, StartedAt: now.Add(-4 * time.Hour), UpdatedAt: now.Add(-4 * time.Hour), Duration: 15 * time.Minute}, // Duration spike
	}

	if err := storage.StoreBatch(ctx, runs); err != nil {
		t.Fatalf("failed to store runs: %v", err)
	}

	anomalies, err := analyzer.DetectAnomalies(ctx, "github", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to detect anomalies: %v", err)
	}

	if len(anomalies) == 0 {
		t.Error("expected to detect anomalies, but none found")
	}

	// Check for high failure rate anomaly
	foundFailureRate := false
	for _, a := range anomalies {
		if a.Type == "high_failure_rate" {
			foundFailureRate = true
			if a.Severity != "warning" {
				t.Errorf("expected severity 'warning', got %s", a.Severity)
			}
		}
	}

	if !foundFailureRate {
		t.Error("expected to find high_failure_rate anomaly")
	}
}

func TestTrendDirection(t *testing.T) {
	tests := []struct {
		direction TrendDirection
		expected  string
	}{
		{TrendUp, "up"},
		{TrendDown, "down"},
		{TrendStable, "stable"},
		{TrendUnknown, "unknown"},
	}

	for _, tt := range tests {
		if got := tt.direction.String(); got != tt.expected {
			t.Errorf("TrendDirection.String() = %s, want %s", got, tt.expected)
		}
	}
}
