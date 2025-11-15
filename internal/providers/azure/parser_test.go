package azure

import (
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestParseAzureRuns(t *testing.T) {
	raw := []byte(`{
        "value": [
            {
                "id": 101,
                "definition": { "name": "CI Pipeline" },
                "sourceBranch": "refs/heads/main",
                "status": "inProgress",
                "result": null,
                "startTime": "2025-11-01T10:00:00Z",
                "finishTime": null,
                "url": "https://dev.azure.com/org/project/_build/results?buildId=101",
                "repository": { "name": "org/project" }
            },
            {
                "id": 100,
                "definition": { "name": "CI Pipeline" },
                "sourceBranch": "refs/heads/main",
                "status": "completed",
                "result": "succeeded",
                "startTime": "2025-10-28T08:00:00Z",
                "finishTime": "2025-10-28T08:07:00Z",
                "url": "https://dev.azure.com/org/project/_build/results?buildId=100",
                "repository": { "name": "org/project" }
            }
        ]
    }`)

	runs, err := ParseAzureRuns(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	inProgress := runs[0]
	if inProgress.Provider != "azure" {
		t.Fatalf("expected provider azure, got %s", inProgress.Provider)
	}
	if inProgress.Status != core.RunStatusInProgress {
		t.Fatalf("expected in progress status, got %s", inProgress.Status)
	}
	if inProgress.Conclusion != "" {
		t.Fatalf("expected empty conclusion for running build, got %s", inProgress.Conclusion)
	}
	if inProgress.Branch != "refs/heads/main" {
		t.Fatalf("expected branch captured, got %s", inProgress.Branch)
	}

	completed := runs[1]
	if completed.Status != core.RunStatusCompleted {
		t.Fatalf("expected completed status, got %s", completed.Status)
	}
	if completed.Conclusion != "succeeded" {
		t.Fatalf("expected succeeded result, got %s", completed.Conclusion)
	}
	startTime, _ := time.Parse(time.RFC3339, "2025-10-28T08:00:00Z")
	finishTime, _ := time.Parse(time.RFC3339, "2025-10-28T08:07:00Z")
	if !completed.StartedAt.Equal(startTime) || !completed.UpdatedAt.Equal(finishTime) {
		t.Fatalf("expected timestamps from API, got %+v", completed)
	}
}
