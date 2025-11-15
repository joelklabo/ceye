package github

import (
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestParseGitHubRuns(t *testing.T) {
	raw := []byte(`{
        "workflow_runs": [
            {
                "id": 1001,
                "name": "Build",
                "head_branch": "main",
                "head_sha": "abcde12345",
                "status": "in_progress",
                "conclusion": null,
                "html_url": "https://github.com/user/repo/actions/runs/1001",
                "created_at": "2025-11-01T12:00:00Z",
                "updated_at": "2025-11-01T12:05:00Z"
            },
            {
                "id": 1000,
                "name": "Build",
                "head_branch": "main",
                "head_sha": "12345abcde",
                "status": "completed",
                "conclusion": "success",
                "html_url": "https://github.com/user/repo/actions/runs/1000",
                "created_at": "2025-10-30T09:00:00Z",
                "updated_at": "2025-10-30T09:10:00Z"
            }
        ]
    }`)

	runs, err := ParseGitHubRuns(raw)
	if err != nil {
		t.Fatalf("expected no error parsing github runs, got %v", err)
	}

	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	first := runs[0]
	if first.Provider != "github" {
		t.Fatalf("expected provider github, got %s", first.Provider)
	}
	if first.Status != core.RunStatusInProgress {
		t.Fatalf("expected in progress status, got %s", first.Status)
	}
	if first.Conclusion != "" {
		t.Fatalf("expected empty conclusion for in-progress run, got %s", first.Conclusion)
	}
	createdAt, _ := time.Parse(time.RFC3339, "2025-11-01T12:00:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2025-11-01T12:05:00Z")
	if !first.StartedAt.Equal(createdAt) || !first.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected timestamps: %+v", first)
	}

	second := runs[1]
	if second.Status != core.RunStatusCompleted {
		t.Fatalf("expected completed status, got %s", second.Status)
	}
	if second.Conclusion != "success" {
		t.Fatalf("expected success conclusion, got %s", second.Conclusion)
	}
	if second.CommitSHA != "12345abcde" {
		t.Fatalf("expected commit sha captured, got %s", second.CommitSHA)
	}
}
