package azure

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

// CLIClient calls the Azure CLI instead of the REST API.
type CLIClient struct{}

// NewCLIClient constructs a CLI-backed Azure client.
func NewCLIClient() *CLIClient {
	return &CLIClient{}
}

func (c *CLIClient) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
	orgURL := normalizeAzureOrgURL(org)
	cmd := exec.Command("az", "pipelines", "runs", "list",
		"--org", orgURL,
		"--project", project,
		"--query", "[].{definitionId:definition.id,definitionName:definition.name,id:id,status:status,result:result,queueTime:queueTime,startTime:startTime,finishTime:finishTime,sourceBranch:sourceBranch,sourceVersion:sourceVersion,webUrl:url}",
		"-o", "json",
	)
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("az pipelines runs list: %w", err)
	}

	var entries []struct {
		DefinitionID   int    `json:"definitionId"`
		DefinitionName string `json:"definitionName"`
		ID             int    `json:"id"`
		Status         string `json:"status"`
		Result         string `json:"result"`
		QueueTime      string `json:"queueTime"`
		StartTime      string `json:"startTime"`
		FinishTime     string `json:"finishTime"`
		SourceBranch   string `json:"sourceBranch"`
		SourceVersion  string `json:"sourceVersion"`
		WebURL         string `json:"webUrl"`
	}
	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, fmt.Errorf("parse az output: %w", err)
	}

	filter := make(map[int]struct{}, len(pipelines))
	for _, id := range pipelines {
		filter[id] = struct{}{}
	}

	runs := make([]core.Run, 0, len(entries))
	for _, entry := range entries {
		if len(filter) > 0 {
			if _, ok := filter[entry.DefinitionID]; !ok {
				continue
			}
		}

		start := parseAzureTime(entry.StartTime)
		if start.IsZero() {
			start = parseAzureTime(entry.QueueTime)
		}
		finish := parseAzureTime(entry.FinishTime)
		duration := finish.Sub(start)
		if duration < 0 {
			duration = 0
		}

		run := core.Run{
			ID:           fmt.Sprintf("%s-%d", entry.DefinitionName, entry.ID),
			Provider:     "azure",
			Repo:         fmt.Sprintf("%s/%s", org, project),
			WorkflowName: entry.DefinitionName,
			Status:       parseAzureStatus(entry.Status),
			Conclusion:   entry.Result,
			Branch:       trimAzureBranch(entry.SourceBranch),
			CommitSHA:    entry.SourceVersion,
			StartedAt:    start,
			UpdatedAt:    finish,
			Duration:     duration,
			URL:          entry.WebURL,
		}
		runs = append(runs, run)
	}

	sort.SliceStable(runs, func(i, j int) bool {
		return runs[i].UpdatedAt.After(runs[j].UpdatedAt)
	})

	return runs, nil
}

func parseAzureStatus(status string) core.RunStatus {
	switch strings.ToLower(status) {
	case "queued":
		return core.RunStatusQueued
	case "inprogress", "in_progress":
		return core.RunStatusInProgress
	case "cancelling", "canceled":
		return core.RunStatusCancelled
	case "completed":
		return core.RunStatusCompleted
	default:
		return core.RunStatusUnknown
	}
}

func parseAzureTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func trimAzureBranch(branch string) string {
	if branch == "" {
		return ""
	}
	return strings.TrimPrefix(branch, "refs/heads/")
}

func normalizeAzureOrgURL(org string) string {
	org = strings.TrimSpace(org)
	org = strings.TrimSuffix(org, "/")
	if strings.HasPrefix(org, "http://") || strings.HasPrefix(org, "https://") {
		return org
	}
	return fmt.Sprintf("https://dev.azure.com/%s", org)
}
