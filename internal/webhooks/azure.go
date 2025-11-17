package webhooks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// AzureWebhookPayload represents Azure DevOps build webhook payload
type AzureWebhookPayload struct {
	EventType string `json:"eventType"`
	Resource  struct {
		ID            int    `json:"id"`
		BuildNumber   string `json:"buildNumber"`
		Status        string `json:"status"`
		Result        string `json:"result"`
		SourceBranch  string `json:"sourceBranch"`
		SourceVersion string `json:"sourceVersion"`
		StartTime     string `json:"startTime"`
		FinishTime    string `json:"finishTime"`
		URL           string `json:"url"`
		Definition    struct {
			Name string `json:"name"`
		} `json:"definition"`
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	} `json:"resource"`
}

// ParseAzureWebhook parses an Azure DevOps build webhook payload into a Run object
func ParseAzureWebhook(data []byte) (core.Run, error) {
	var payload AzureWebhookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return core.Run{}, fmt.Errorf("parse azure webhook: %w", err)
	}

	// Parse timestamps
	startedAt, err := time.Parse(time.RFC3339, payload.Resource.StartTime)
	if err != nil {
		return core.Run{}, fmt.Errorf("parse start time: %w", err)
	}
	
	// FinishTime might be empty for in-progress builds
	var updatedAt time.Time
	if payload.Resource.FinishTime != "" {
		updatedAt, err = time.Parse(time.RFC3339, payload.Resource.FinishTime)
		if err != nil {
			return core.Run{}, fmt.Errorf("parse finish time: %w", err)
		}
	} else {
		updatedAt = time.Now()
	}

	// Map status
	status := mapAzureStatus(payload.Resource.Status)
	
	// Get result/conclusion
	conclusion := payload.Resource.Result

	// Clean branch name (Azure returns "refs/heads/main")
	branch := strings.TrimPrefix(payload.Resource.SourceBranch, "refs/heads/")

	// Repo is the project name in Azure DevOps
	repo := payload.Resource.Project.Name

	run := core.Run{
		ID:           fmt.Sprintf("%d", payload.Resource.ID),
		Provider:     "azure",
		Repo:         repo,
		WorkflowName: payload.Resource.Definition.Name,
		Branch:       branch,
		CommitSHA:    payload.Resource.SourceVersion,
		Status:       status,
		Conclusion:   conclusion,
		StartedAt:    startedAt,
		UpdatedAt:    updatedAt,
		Duration:     updatedAt.Sub(startedAt),
		URL:          payload.Resource.URL,
	}

	return run, nil
}

// mapAzureStatus converts Azure DevOps status strings to core.RunStatus
func mapAzureStatus(raw string) core.RunStatus {
	switch strings.ToLower(raw) {
	case "notstarted", "postponed":
		return core.RunStatusQueued
	case "inprogress":
		return core.RunStatusInProgress
	case "completed":
		return core.RunStatusCompleted
	case "cancelling", "cancelled":
		return core.RunStatusCancelled
	default:
		return core.RunStatusUnknown
	}
}
