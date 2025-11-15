package github

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

type stubGitHubClient struct {
	runs []core.Run
}

func (s *stubGitHubClient) ListWorkflowRuns(owner, repo string) ([]core.Run, error) {
	return s.runs, nil
}

func TestProviderStartEmitsEventsAndStopsOnContextCancel(t *testing.T) {
	runs := []core.Run{
		{ID: "1", Provider: "github", WorkflowName: "Build", Status: core.RunStatusInProgress},
	}
	client := &stubGitHubClient{runs: runs}
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
