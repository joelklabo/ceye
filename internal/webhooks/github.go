package webhooks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// GitHubWebhookPayload represents the GitHub workflow_run webhook payload
type GitHubWebhookPayload struct {
	Action      string `json:"action"`
	WorkflowRun struct {
		ID         int64   `json:"id"`
		Name       string  `json:"name"`
		HeadBranch string  `json:"head_branch"`
		HeadSHA    string  `json:"head_sha"`
		Status     string  `json:"status"`
		Conclusion *string `json:"conclusion"`
		HTMLURL    string  `json:"html_url"`
		CreatedAt  string  `json:"created_at"`
		UpdatedAt  string  `json:"updated_at"`
	} `json:"workflow_run"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// ParseGitHubWebhook parses a GitHub workflow_run webhook payload into a Run object
func ParseGitHubWebhook(data []byte) (core.Run, error) {
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return core.Run{}, fmt.Errorf("parse github webhook: %w", err)
	}

	// Parse timestamps
	startedAt, err := time.Parse(time.RFC3339, payload.WorkflowRun.CreatedAt)
	if err != nil {
		return core.Run{}, fmt.Errorf("parse start time: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, payload.WorkflowRun.UpdatedAt)
	if err != nil {
		return core.Run{}, fmt.Errorf("parse update time: %w", err)
	}

	// Map status
	status := mapGitHubStatus(payload.WorkflowRun.Status)
	
	// Get conclusion if available
	conclusion := ""
	if payload.WorkflowRun.Conclusion != nil {
		conclusion = *payload.WorkflowRun.Conclusion
	}

	// Extract repo from full_name (format: "owner/repo")
	repo := payload.Repository.FullName

	run := core.Run{
		ID:           fmt.Sprintf("%d", payload.WorkflowRun.ID),
		Provider:     "github",
		Repo:         repo,
		WorkflowName: payload.WorkflowRun.Name,
		Branch:       payload.WorkflowRun.HeadBranch,
		CommitSHA:    payload.WorkflowRun.HeadSHA,
		Status:       status,
		Conclusion:   conclusion,
		StartedAt:    startedAt,
		UpdatedAt:    updatedAt,
		Duration:     updatedAt.Sub(startedAt),
		URL:          payload.WorkflowRun.HTMLURL,
	}

	return run, nil
}

// mapGitHubStatus converts GitHub status strings to core.RunStatus
func mapGitHubStatus(raw string) core.RunStatus {
	switch strings.ToLower(raw) {
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
