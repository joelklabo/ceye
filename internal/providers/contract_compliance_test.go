package providers

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	demoprovider "github.com/joelklabo/ceye/internal/providers/demo"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
)

// TestDemoProviderContract validates demo provider against basic contract requirements
func TestDemoProviderContract(t *testing.T) {
	provider := demoprovider.New(2, 100*time.Millisecond)

	// Test Name
	if provider.Name() == "" {
		t.Error("Provider.Name() returned empty string")
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	eventCh := make(chan core.RunEvent, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- provider.Start(ctx, eventCh)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-errCh:
		t.Log("✓ Demo provider respects context cancellation")
	case <-time.After(2 * time.Second):
		t.Error("Demo provider did not respect context cancellation")
	}
}

// TestGitHubProviderContract validates GitHub provider against basic contract
func TestGitHubProviderContract(t *testing.T) {
	mockClient := &mockGitHubClient{}
	provider := githubprovider.NewProvider(mockClient, []githubprovider.RepoConfig{
		{Owner: "test", Repo: "repo"},
	})

	// Test Name
	if provider.Name() != "github" {
		t.Errorf("expected name 'github', got %q", provider.Name())
	}

	// Test that it sends events
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	eventCh := make(chan core.RunEvent, 10)
	go provider.Start(ctx, eventCh)

	select {
	case event := <-eventCh:
		if event.Provider != "github" {
			t.Errorf("event.Provider = %q, want github", event.Provider)
		}
		t.Log("✓ GitHub provider sends events")
	case <-time.After(1 * time.Second):
		t.Error("GitHub provider did not send event within timeout")
	}
}

// TestSafeProviderContract validates SafeProvider wrapper maintains contract
func TestSafeProviderContract(t *testing.T) {
	inner := demoprovider.New(2, 100*time.Millisecond)
	provider := NewSafeProvider(inner)

	// Test Name passes through
	if provider.Name() != inner.Name() {
		t.Error("SafeProvider.Name() does not match inner provider")
	}

	// Test that events flow through
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	eventCh := make(chan core.RunEvent, 10)
	go provider.Start(ctx, eventCh)

	select {
	case event := <-eventCh:
		if event.Provider != "demo" {
			t.Errorf("event.Provider = %q, want demo", event.Provider)
		}
		t.Log("✓ SafeProvider forwards events")
	case <-time.After(1 * time.Second):
		t.Error("SafeProvider did not forward event")
	}
}

// mockGitHubClient for testing
type mockGitHubClient struct{}

func (m *mockGitHubClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	return []core.Run{
		{
			ID:           "test-run-1",
			Provider:     "github",
			Repo:         owner + "/" + repo,
			WorkflowName: "CI",
			Status:       core.RunStatusInProgress,
		},
	}, nil
}

// mockAzureClient for testing
type mockAzureClient struct{}

func (m *mockAzureClient) ListPipelineRuns(org, project string, pipelineID int) ([]core.Run, error) {
	return []core.Run{
		{
			ID:           "test-run-1",
			Provider:     "azure",
			Repo:         project,
			WorkflowName: "Build",
			Status:       core.RunStatusInProgress,
		},
	}, nil
}
