package azure

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

type stubAzureClient struct {
	responses [][]core.Run
}

func newStubAzureClient(responses ...[]core.Run) *stubAzureClient {
	return &stubAzureClient{responses: responses}
}

func (s *stubAzureClient) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
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
	runs := []core.Run{{ID: "42", Provider: "azure", WorkflowName: "Build", Status: core.RunStatusInProgress}}
	client := newStubAzureClient(runs)
	provider := NewProvider(client, Config{Org: "org", Project: "proj", Pipelines: []int{1}})

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
		if evt.Provider != "azure" {
			t.Fatalf("expected provider azure, got %s", evt.Provider)
		}
		if len(evt.Runs) != 1 {
			t.Fatalf("expected 1 run in event, got %d", len(evt.Runs))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for azure event")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("azure provider did not exit after context cancel")
	}
}

func TestAzureProviderRefresh(t *testing.T) {
	first := []core.Run{{ID: "1", Provider: "azure", WorkflowName: "Build", Status: core.RunStatusCompleted}}
	second := []core.Run{{ID: "2", Provider: "azure", WorkflowName: "Deploy", Status: core.RunStatusCompleted}}
	client := newStubAzureClient(first, second)
	provider := NewProvider(client, Config{Org: "org", Project: "proj", Pipelines: []int{1}})
	provider.fastInterval = time.Hour
	provider.slowInterval = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan core.RunEvent, 2)

	go provider.Start(ctx, events)

	select {
	case <-events:
	case <-time.After(2 * time.Second):
		t.Fatal("expected first azure event")
	}

	provider.TriggerRefresh()

	select {
	case evt := <-events:
		if len(evt.Runs) == 0 || evt.Runs[0].ID != "2" {
			t.Fatalf("expected refreshed azure run, got %+v", evt.Runs)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected azure refresh to trigger poll")
	}
}
