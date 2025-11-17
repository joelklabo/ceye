package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// ParseGitHubRuns converts GitHub workflow runs JSON into core.Run structures.
func ParseGitHubRuns(data []byte) ([]core.Run, error) {
	var payload struct {
		WorkflowRuns []struct {
			ID         int64   `json:"id"`
			Name       string  `json:"name"`
			HeadBranch string  `json:"head_branch"`
			HeadSHA    string  `json:"head_sha"`
			Status     string  `json:"status"`
			Conclusion *string `json:"conclusion"`
			HTMLURL    string  `json:"html_url"`
			CreatedAt  string  `json:"created_at"`
			UpdatedAt  string  `json:"updated_at"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"workflow_runs"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse github runs: %w", err)
	}

	runs := make([]core.Run, 0, len(payload.WorkflowRuns))
	for _, wr := range payload.WorkflowRuns {
		startedAt, err := time.Parse(time.RFC3339, wr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse start time: %w", err)
		}
		updatedAt, err := time.Parse(time.RFC3339, wr.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse update time: %w", err)
		}

		status := mapGitHubStatus(wr.Status)
		conclusion := ""
		if wr.Conclusion != nil {
			conclusion = *wr.Conclusion
		}

		runs = append(runs, core.Run{
			ID:           fmt.Sprintf("%d", wr.ID),
			Provider:     "github",
			Repo:         wr.Repository.FullName,
			WorkflowName: wr.Name,
			Branch:       wr.HeadBranch,
			CommitSHA:    wr.HeadSHA,
			Status:       status,
			Conclusion:   conclusion,
			StartedAt:    startedAt,
			UpdatedAt:    updatedAt,
			Duration:     updatedAt.Sub(startedAt),
			URL:          wr.HTMLURL,
		})
	}

	return runs, nil
}

func mapGitHubStatus(raw string) core.RunStatus {
	switch raw {
	case "queued":
		return core.RunStatusQueued
	case "in_progress":
		return core.RunStatusInProgress
	case "completed":
		return core.RunStatusCompleted
	default:
		return core.RunStatusUnknown
	}
}
