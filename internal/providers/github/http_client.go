package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// HTTPClient implements GitHubClient using the REST API.
type HTTPClient struct {
	httpClient *http.Client
	token      string
}

// NewHTTPClient constructs a GitHub HTTP client using the provided auth token.
func NewHTTPClient(token string) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		token:      token,
	}
}

func (c *HTTPClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs", owner, repo), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github api: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseGitHubRuns(body)
}
