package github

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

type stubGitHubClient struct {
	responses [][]core.Run
}

func newStubGitHubClient(responses ...[]core.Run) *stubGitHubClient {
	return &stubGitHubClient{responses: responses}
}

func (s *stubGitHubClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	if len(s.responses) == 0 {
		return nil, nil
	}
	resp := s.responses[0]
	if len(s.responses) > 1 {
		s.responses = s.responses[1:]
	}
	return resp, nil
}

func TestProviderStartEmitsEventsAndStopsOnContextCancel(t *testing.T) {
	runs := []core.Run{
		{ID: "1", Provider: "github", WorkflowName: "Build", Status: core.RunStatusInProgress},
	}
	client := newStubGitHubClient(runs)
	provider := NewProvider(client, []RepoConfig{{Owner: "user", Repo: "repo"}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan core.RunEvent, 1)

	done := make(chan struct{})
	go func() {
		if err := provider.Start(ctx, events); err != nil && err != context.Canceled {
			t.Errorf("unexpected error from Start: %v", err)
		}
		close(done)
	}()

	select {
	case evt := <-events:
		if evt.Provider != "github" {
			t.Fatalf("expected provider github, got %s", evt.Provider)
		}
		if len(evt.Runs) != 1 {
			t.Fatalf("expected 1 run in event, got %d", len(evt.Runs))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event from provider")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("provider did not exit after context cancel")
	}
}

func TestProviderTriggerRefreshForcesPoll(t *testing.T) {
	first := []core.Run{{ID: "1", Provider: "github", WorkflowName: "Build", Status: core.RunStatusCompleted}}
	second := []core.Run{{ID: "2", Provider: "github", WorkflowName: "Test", Status: core.RunStatusCompleted}}
	client := newStubGitHubClient(first, second)
	provider := NewProvider(client, []RepoConfig{{Owner: "user", Repo: "repo"}})
	provider.fastInterval = time.Hour
	provider.slowInterval = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan core.RunEvent, 2)

	go provider.Start(ctx, events)

	select {
	case <-events:
	case <-time.After(2 * time.Second):
		t.Fatal("expected first event without refresh")
	}

	provider.TriggerRefresh()

	select {
	case evt := <-events:
		if len(evt.Runs) == 0 || evt.Runs[0].ID != "2" {
			t.Fatalf("expected refreshed data, got %+v", evt.Runs)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected refresh to trigger immediate poll")
	}
}
