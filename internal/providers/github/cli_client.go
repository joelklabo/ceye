package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// CLIClient calls the GitHub CLI instead of the REST API.
type CLIClient struct{}

// NewCLIClient constructs a CLI-backed GitHub client.
func NewCLIClient() *CLIClient {
	return &CLIClient{}
}

func (c *CLIClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	cmd := exec.Command("gh", "run", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--limit", "25",
		"--json", "number,status,conclusion,workflowName,headBranch,headSha,updatedAt,startedAt,url",
	)
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list: %w", err)
	}

	var entries []struct {
		Number       int    `json:"number"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		WorkflowName string `json:"workflowName"`
		HeadBranch   string `json:"headBranch"`
		HeadSha      string `json:"headSha"`
		UpdatedAt    string `json:"updatedAt"`
		StartedAt    string `json:"startedAt"`
		URL          string `json:"url"`
	}
	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}

	runs := make([]core.Run, 0, len(entries))
	for _, entry := range entries {
		started, _ := time.Parse(time.RFC3339, entry.StartedAt)
		updated, _ := time.Parse(time.RFC3339, entry.UpdatedAt)
		duration := updated.Sub(started)
		if duration < 0 {
			duration = 0
		}
		run := core.Run{
			ID:           fmt.Sprintf("%s#%d", repo, entry.Number),
			Provider:     "github",
			Repo:         repo,
			WorkflowName: entry.WorkflowName,
			Status:       parseGHStatus(entry.Status),
			Conclusion:   entry.Conclusion,
			Branch:       entry.HeadBranch,
			CommitSHA:    entry.HeadSha,
			StartedAt:    started,
			UpdatedAt:    updated,
			Duration:     duration,
			URL:          entry.URL,
		}
		runs = append(runs, run)
	}

	sort.SliceStable(runs, func(i, j int) bool {
		return runs[i].UpdatedAt.After(runs[j].UpdatedAt)
	})
	return runs, nil
}

func parseGHStatus(status string) core.RunStatus {
	switch strings.ToLower(status) {
	case "queued", "requested", "waiting", "pending":
		return core.RunStatusQueued
	case "in_progress":
		return core.RunStatusInProgress
	case "completed":
		return core.RunStatusCompleted
	default:
		return core.RunStatusUnknown
	}
}
