package azure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

const defaultAPIVersion = "7.1-preview.6"

// HTTPClient implements AzureClient via the REST API.
type HTTPClient struct {
	httpClient *http.Client
	pat        string
}

// NewHTTPClient builds an Azure client using a PAT.
func NewHTTPClient(pat string) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		pat:        pat,
	}
}

func (c *HTTPClient) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
	if org == "" || project == "" {
		return nil, fmt.Errorf("org and project required")
	}
	apiURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/build/builds?api-version=%s", url.PathEscape(org), url.PathEscape(project), defaultAPIVersion)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	if c.pat != "" {
		req.SetBasicAuth("", c.pat)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseAzureRuns(data)
}
