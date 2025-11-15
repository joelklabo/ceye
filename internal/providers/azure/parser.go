package azure

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// ParseAzureRuns converts Azure DevOps build JSON into core.Run structures.
func ParseAzureRuns(data []byte) ([]core.Run, error) {
	var payload struct {
		Value []struct {
			ID         int64 `json:"id"`
			Definition struct {
				Name string `json:"name"`
			} `json:"definition"`
			SourceBranch string  `json:"sourceBranch"`
			Status       string  `json:"status"`
			Result       *string `json:"result"`
			StartTime    string  `json:"startTime"`
			FinishTime   *string `json:"finishTime"`
			URL          string  `json:"url"`
			Repository   struct {
				Name string `json:"name"`
			} `json:"repository"`
		} `json:"value"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse azure runs: %w", err)
	}

	runs := make([]core.Run, 0, len(payload.Value))
	for _, build := range payload.Value {
		start, err := time.Parse(time.RFC3339, build.StartTime)
		if err != nil {
			return nil, fmt.Errorf("parse start time: %w", err)
		}

		var updated time.Time
		if build.FinishTime != nil {
			updated, err = time.Parse(time.RFC3339, *build.FinishTime)
			if err != nil {
				return nil, fmt.Errorf("parse finish time: %w", err)
			}
		} else {
			updated = start
		}

		result := ""
		if build.Result != nil {
			result = *build.Result
		}

		runs = append(runs, core.Run{
			ID:           fmt.Sprintf("%d", build.ID),
			Provider:     "azure",
			Repo:         build.Repository.Name,
			WorkflowName: build.Definition.Name,
			Status:       mapAzureStatus(build.Status),
			Conclusion:   result,
			Branch:       build.SourceBranch,
			StartedAt:    start,
			UpdatedAt:    updated,
			Duration:     updated.Sub(start),
			URL:          build.URL,
		})
	}

	return runs, nil
}

func mapAzureStatus(raw string) core.RunStatus {
	switch strings.ToLower(raw) {
	case "queued", "notstarted":
		return core.RunStatusQueued
	case "inprogress":
		return core.RunStatusInProgress
	case "completed":
		return core.RunStatusCompleted
	default:
		return core.RunStatusUnknown
	}
}
