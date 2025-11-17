package github

import (
	"testing"
	"time"
)

func TestDurationCalculation(t *testing.T) {
	// Test JSON with a run that has a 2 minute duration
	jsonData := `{
		"workflow_runs": [{
			"id": 123,
			"name": "CI",
			"head_branch": "main",
			"head_sha": "abc123",
			"status": "completed",
			"conclusion": "success",
			"html_url": "https://github.com/test/repo/actions/runs/123",
			"created_at": "2025-11-17T10:00:00Z",
			"updated_at": "2025-11-17T10:02:34Z",
			"repository": {
				"full_name": "test/repo"
			}
		}]
	}`

	runs, err := ParseGitHubRuns([]byte(jsonData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	
	// Duration should be 2m 34s = 154 seconds
	expectedDuration := 2*time.Minute + 34*time.Second
	if run.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v (got %d seconds, want %d seconds)", 
			run.Duration, expectedDuration, 
			int(run.Duration.Seconds()), 
			int(expectedDuration.Seconds()))
	}

	// Verify Repo field is set
	if run.Repo != "test/repo" {
		t.Errorf("Repo = %q, want %q", run.Repo, "test/repo")
	}

	t.Logf("✅ Duration correctly calculated: %v (%d seconds)", run.Duration, int(run.Duration.Seconds()))
	t.Logf("✅ Repo correctly parsed: %s", run.Repo)
}

func TestDurationForInProgressRun(t *testing.T) {
	// Test JSON with a run that's in progress (duration will be elapsed time)
	jsonData := `{
		"workflow_runs": [{
			"id": 456,
			"name": "Build",
			"head_branch": "feature",
			"head_sha": "def456",
			"status": "in_progress",
			"conclusion": null,
			"html_url": "https://github.com/test/repo/actions/runs/456",
			"created_at": "2025-11-17T10:00:00Z",
			"updated_at": "2025-11-17T10:01:15Z",
			"repository": {
				"full_name": "test/repo"
			}
		}]
	}`

	runs, err := ParseGitHubRuns([]byte(jsonData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	run := runs[0]
	
	// Duration should be 1m 15s = 75 seconds (elapsed time)
	expectedDuration := 1*time.Minute + 15*time.Second
	if run.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", run.Duration, expectedDuration)
	}

	t.Logf("✅ In-progress run duration correctly calculated: %v (%d seconds)", run.Duration, int(run.Duration.Seconds()))
}

func TestDurationForFastRun(t *testing.T) {
	// Test JSON with a very fast run (10 seconds)
	jsonData := `{
		"workflow_runs": [{
			"id": 789,
			"name": "Lint",
			"head_branch": "main",
			"head_sha": "ghi789",
			"status": "completed",
			"conclusion": "success",
			"html_url": "https://github.com/test/repo/actions/runs/789",
			"created_at": "2025-11-17T10:00:00Z",
			"updated_at": "2025-11-17T10:00:10Z",
			"repository": {
				"full_name": "test/repo"
			}
		}]
	}`

	runs, err := ParseGitHubRuns([]byte(jsonData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	run := runs[0]
	
	// Duration should be 10 seconds
	expectedDuration := 10 * time.Second
	if run.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", run.Duration, expectedDuration)
	}

	t.Logf("✅ Fast run duration correctly calculated: %v (%d seconds)", run.Duration, int(run.Duration.Seconds()))
}
