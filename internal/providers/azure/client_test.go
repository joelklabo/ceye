package azure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestClientListBuilds(t *testing.T) {
	builds := []AzureBuild{
		{
			ID:          123,
			BuildNumber: "20230101.1",
			Status:      "completed",
			Result:      "succeeded",
			StartTime:   "2023-01-01T10:00:00Z",
			FinishTime:  "2023-01-01T10:05:00Z",
			Definition: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Path string `json:"path"`
			}{
				ID:   1,
				Name: "Build",
			},
			Project: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				Name: "MyProject",
			},
			Repository: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			}{
				Name: "MyRepo",
			},
			SourceBranch:  "refs/heads/main",
			SourceVersion: "abc123",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Check authentication
		user, pass, ok := r.BasicAuth()
		if !ok || user != "" || pass != "test-pat" {
			t.Error("expected basic auth with PAT")
		}

		response := struct {
			Value []AzureBuild `json:"value"`
			Count int          `json:"count"`
		}{
			Value: builds,
			Count: len(builds),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom server
	client := NewClient("test-org", "test-pat")
	client.httpClient = server.Client()

	// Override URL for testing (normally would need to change the method)
	// For now, we'll test via the actual Azure URL format
	t.Skip("Need to refactor client to accept base URL for testing")
}

func TestMapAzureStatus(t *testing.T) {
	tests := []struct {
		status   string
		result   string
		expected core.RunStatus
	}{
		{"notStarted", "", core.RunStatusQueued},
		{"inProgress", "", core.RunStatusInProgress},
		{"completed", "succeeded", core.RunStatusCompleted},
		{"completed", "failed", core.RunStatusFailed},
		{"completed", "canceled", core.RunStatusCancelled},
		{"completed", "partiallySucceeded", core.RunStatusCompleted},
		{"cancelling", "", core.RunStatusCancelled},
		{"postponed", "", core.RunStatusQueued},
		{"unknown", "", core.RunStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.status, tt.result), func(t *testing.T) {
			got := mapAzureStatus(tt.status, tt.result)
			if got != tt.expected {
				t.Errorf("mapAzureStatus(%q, %q) = %v, want %v",
					tt.status, tt.result, got, tt.expected)
			}
		})
	}
}

func TestMapAzureConclusion(t *testing.T) {
	tests := []struct {
		result   string
		expected string
	}{
		{"succeeded", "success"},
		{"partiallySucceeded", "partial_success"},
		{"failed", "failure"},
		{"canceled", "cancelled"},
		{"cancelled", "cancelled"},
		{"", ""},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			got := mapAzureConclusion(tt.result)
			if got != tt.expected {
				t.Errorf("mapAzureConclusion(%q) = %q, want %q",
					tt.result, got, tt.expected)
			}
		})
	}
}

func TestCleanBranchName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"refs/heads/main", "main"},
		{"refs/heads/feature/test", "feature/test"},
		{"refs/tags/v1.0.0", "v1.0.0"},
		{"refs/pull/123", "123"},
		{"main", "main"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanBranchName(tt.input)
			if got != tt.expected {
				t.Errorf("cleanBranchName(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseAzureBuilds(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	later := time.Now().Add(5 * time.Minute).Format(time.RFC3339)

	builds := []AzureBuild{
		{
			ID:          123,
			BuildNumber: "20230101.1",
			Status:      "completed",
			Result:      "succeeded",
			StartTime:   now,
			FinishTime:  later,
			Definition: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Path string `json:"path"`
			}{
				Name: "Build",
			},
			Project: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				Name: "MyProject",
			},
			Repository: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			}{
				Name: "MyRepo",
			},
			SourceBranch:  "refs/heads/main",
			SourceVersion: "abc123",
			Links: struct {
				Web struct {
					Href string `json:"href"`
				} `json:"web"`
				Timeline struct {
					Href string `json:"href"`
				} `json:"timeline"`
			}{
				Web: struct {
					Href string `json:"href"`
				}{
					Href: "https://dev.azure.com/org/project/_build/results?buildId=123",
				},
			},
		},
	}

	runs := parseAzureBuilds(builds)

	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]

	if run.ID != "123" {
		t.Errorf("ID = %q, want %q", run.ID, "123")
	}

	if run.Provider != "azure" {
		t.Errorf("Provider = %q, want %q", run.Provider, "azure")
	}

	if run.Repo != "MyProject/MyRepo" {
		t.Errorf("Repo = %q, want %q", run.Repo, "MyProject/MyRepo")
	}

	if run.WorkflowName != "Build" {
		t.Errorf("WorkflowName = %q, want %q", run.WorkflowName, "Build")
	}

	if run.Status != core.RunStatusCompleted {
		t.Errorf("Status = %v, want %v", run.Status, core.RunStatusCompleted)
	}

	if run.Conclusion != "success" {
		t.Errorf("Conclusion = %q, want %q", run.Conclusion, "success")
	}

	if run.Branch != "main" {
		t.Errorf("Branch = %q, want %q", run.Branch, "main")
	}

	if run.CommitSHA != "abc123" {
		t.Errorf("CommitSHA = %q, want %q", run.CommitSHA, "abc123")
	}

	if run.URL == "" {
		t.Error("URL should not be empty")
	}

	if run.Duration == 0 {
		t.Error("Duration should be calculated")
	}
}

func TestParseAzureBuildsInProgress(t *testing.T) {
	now := time.Now().Format(time.RFC3339)

	builds := []AzureBuild{
		{
			ID:         456,
			Status:     "inProgress",
			Result:     "",
			StartTime:  now,
			FinishTime: "", // No finish time for in-progress builds
			Definition: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Path string `json:"path"`
			}{
				Name: "Deploy",
			},
			Project: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				Name: "Prod",
			},
			Repository: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			}{
				Name: "api",
			},
			SourceBranch:  "refs/heads/release",
			SourceVersion: "def456",
		},
	}

	runs := parseAzureBuilds(builds)

	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]

	if run.Status != core.RunStatusInProgress {
		t.Errorf("Status = %v, want %v", run.Status, core.RunStatusInProgress)
	}

	if run.Conclusion != "" {
		t.Errorf("Conclusion = %q, want empty string", run.Conclusion)
	}

	if run.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set to current time for in-progress builds")
	}

	if run.Duration == 0 {
		t.Error("Duration should be calculated even for in-progress builds")
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		response := struct {
			Value []AzureBuild `json:"value"`
			Count int          `json:"count"`
		}{
			Value: []AzureBuild{},
			Count: 0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Skip("Need to refactor client to accept base URL for testing")
	
	// Would test that client retries on 503
	// Would verify it eventually succeeds
	// Would check retry delay
}

func TestClientRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	t.Skip("Need to refactor client to accept base URL for testing")
	
	// Would test that client handles rate limiting
	// Would verify it waits for Retry-After duration
	// Would check it eventually fails after max retries
}
