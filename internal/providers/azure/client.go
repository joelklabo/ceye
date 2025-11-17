package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

const (
	defaultAPIVersion = "7.1"
	defaultTimeout    = 30 * time.Second
	maxRetries        = 3
	retryDelay        = time.Second
)

// Client implements comprehensive Azure DevOps REST API access
type Client struct {
	httpClient *http.Client
	pat        string
	org        string
	maxRetries int
}

// NewClient creates a new Azure DevOps API client
func NewClient(org, pat string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		pat:        pat,
		org:        org,
		maxRetries: maxRetries,
	}
}

// ListBuilds fetches builds for a project with optional pipeline filtering
func (c *Client) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
	if org == "" {
		org = c.org
	}
	if org == "" || project == "" {
		return nil, fmt.Errorf("org and project required")
	}

	params := url.Values{}
	params.Set("api-version", defaultAPIVersion)
	params.Set("$top", "50") // Limit to recent builds
	
	// Filter by pipeline IDs if provided
	if len(pipelines) > 0 {
		definitions := make([]string, len(pipelines))
		for i, id := range pipelines {
			definitions[i] = strconv.Itoa(id)
		}
		params.Set("definitions", strings.Join(definitions, ","))
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/builds?%s",
		url.PathEscape(org),
		url.PathEscape(project),
		params.Encode())

	var response struct {
		Value []AzureBuild `json:"value"`
		Count int          `json:"count"`
	}

	if err := c.doRequest(apiURL, &response); err != nil {
		return nil, fmt.Errorf("list builds: %w", err)
	}

	return parseAzureBuilds(response.Value), nil
}

// GetBuildDetails fetches detailed information for a specific build
func (c *Client) GetBuildDetails(org, project string, buildID int) (*AzureBuild, error) {
	if org == "" {
		org = c.org
	}
	if org == "" || project == "" {
		return nil, fmt.Errorf("org and project required")
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/builds/%d?api-version=%s",
		url.PathEscape(org),
		url.PathEscape(project),
		buildID,
		defaultAPIVersion)

	var build AzureBuild
	if err := c.doRequest(apiURL, &build); err != nil {
		return nil, fmt.Errorf("get build details: %w", err)
	}

	return &build, nil
}

// ListPipelines fetches all pipeline definitions for a project
func (c *Client) ListPipelines(org, project string) ([]Pipeline, error) {
	if org == "" {
		org = c.org
	}
	if org == "" || project == "" {
		return nil, fmt.Errorf("org and project required")
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/pipelines?api-version=%s",
		url.PathEscape(org),
		url.PathEscape(project),
		defaultAPIVersion)

	var response struct {
		Value []Pipeline `json:"value"`
		Count int        `json:"count"`
	}

	if err := c.doRequest(apiURL, &response); err != nil {
		return nil, fmt.Errorf("list pipelines: %w", err)
	}

	return response.Value, nil
}

// GetBuildTimeline fetches timeline/stages for a build
func (c *Client) GetBuildTimeline(org, project string, buildID int) (*Timeline, error) {
	if org == "" {
		org = c.org
	}
	if org == "" || project == "" {
		return nil, fmt.Errorf("org and project required")
	}

	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/builds/%d/timeline?api-version=%s",
		url.PathEscape(org),
		url.PathEscape(project),
		buildID,
		defaultAPIVersion)

	var timeline Timeline
	if err := c.doRequest(apiURL, &timeline); err != nil {
		return nil, fmt.Errorf("get build timeline: %w", err)
	}

	return &timeline, nil
}

// doRequest executes an HTTP request with retries and error handling
func (c *Client) doRequest(apiURL string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		// Set authentication
		if c.pat != "" {
			req.SetBasicAuth("", c.pat)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited")
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(seconds) * time.Second)
				}
			}
			continue
		}

		// Handle errors
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("api error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		// Parse response
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			continue
		}

		if err := json.Unmarshal(data, result); err != nil {
			// Log first 200 chars of response for debugging
			preview := string(data)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return fmt.Errorf("parse response: %w (response preview: %s)", err, preview)
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", c.maxRetries, lastErr)
}

// AzureBuild represents a build from Azure DevOps API
type AzureBuild struct {
	ID         int64  `json:"id"`
	BuildNumber string `json:"buildNumber"`
	Status     string `json:"status"`
	Result     string `json:"result"`
	QueueTime  string `json:"queueTime"`
	StartTime  string `json:"startTime"`
	FinishTime string `json:"finishTime"`
	
	Definition struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"definition"`
	
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	
	Repository struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"repository"`
	
	SourceBranch  string `json:"sourceBranch"`
	SourceVersion string `json:"sourceVersion"`
	
	Links struct {
		Web struct {
			Href string `json:"href"`
		} `json:"web"`
		Timeline struct {
			Href string `json:"href"`
		} `json:"timeline"`
	} `json:"_links"`
	
	URL string `json:"url"`
}

// Pipeline represents a pipeline definition
type Pipeline struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Folder string `json:"folder"`
	
	Links struct {
		Web struct {
			Href string `json:"href"`
		} `json:"web"`
	} `json:"_links"`
}

// Timeline represents a build timeline with stages/jobs
type Timeline struct {
	ID string `json:"id"`
	Records []TimelineRecord `json:"records"`
}

// TimelineRecord represents a stage, job, or task in a build
type TimelineRecord struct {
	ID            string `json:"id"`
	ParentID      string `json:"parentId"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	StartTime     string `json:"startTime"`
	FinishTime    string `json:"finishTime"`
	State         string `json:"state"`
	Result        string `json:"result"`
	ErrorCount    int    `json:"errorCount"`
	WarningCount  int    `json:"warningCount"`
	Order         int    `json:"order"`
}

// parseAzureBuilds converts Azure API builds to core.Run
func parseAzureBuilds(builds []AzureBuild) []core.Run {
	runs := make([]core.Run, 0, len(builds))
	
	for _, build := range builds {
		run := core.Run{
			ID:           fmt.Sprintf("%d", build.ID),
			Provider:     "azure",
			Repo:         fmt.Sprintf("%s/%s", build.Project.Name, build.Repository.Name),
			WorkflowName: build.Definition.Name,
			Status:       mapAzureStatus(build.Status, build.Result),
			Conclusion:   mapAzureConclusion(build.Result),
			Branch:       cleanBranchName(build.SourceBranch),
			CommitSHA:    build.SourceVersion,
			URL:          build.Links.Web.Href,
		}

		// Parse times
		if build.StartTime != "" {
			if t, err := time.Parse(time.RFC3339, build.StartTime); err == nil {
				run.StartedAt = t
			}
		}

		if build.FinishTime != "" {
			if t, err := time.Parse(time.RFC3339, build.FinishTime); err == nil {
				run.UpdatedAt = t
				run.Duration = run.UpdatedAt.Sub(run.StartedAt)
			}
		} else {
			run.UpdatedAt = time.Now()
			if !run.StartedAt.IsZero() {
				run.Duration = time.Since(run.StartedAt)
			}
		}

		runs = append(runs, run)
	}
	
	return runs
}

// mapAzureStatus maps Azure build status to core.RunStatus
func mapAzureStatus(status, result string) core.RunStatus {
	status = strings.ToLower(status)
	result = strings.ToLower(result)

	switch status {
	case "notstarted":
		return core.RunStatusQueued
	case "inprogress":
		return core.RunStatusInProgress
	case "completed":
		switch result {
		case "succeeded", "partiallysucceeded":
			return core.RunStatusCompleted
		case "failed":
			return core.RunStatusFailed
		case "canceled", "cancelled":
			return core.RunStatusCancelled
		default:
			return core.RunStatusCompleted
		}
	case "cancelling":
		return core.RunStatusCancelled
	case "postponed":
		return core.RunStatusQueued
	default:
		return core.RunStatusUnknown
	}
}

// mapAzureConclusion maps Azure result to a conclusion string
func mapAzureConclusion(result string) string {
	if result == "" {
		return ""
	}
	result = strings.ToLower(result)
	switch result {
	case "succeeded":
		return "success"
	case "partiallysucceeded":
		return "partial_success"
	case "failed":
		return "failure"
	case "canceled", "cancelled":
		return "cancelled"
	default:
		return result
	}
}

// cleanBranchName removes refs/ prefix from branch names
func cleanBranchName(branch string) string {
	branch = strings.TrimPrefix(branch, "refs/heads/")
	branch = strings.TrimPrefix(branch, "refs/tags/")
	branch = strings.TrimPrefix(branch, "refs/pull/")
	return branch
}
