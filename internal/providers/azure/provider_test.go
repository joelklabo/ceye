package azure

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

type stubAzureClient struct {
	runs []core.Run
}

func (s *stubAzureClient) ListBuilds(org, project string, pipelines []int) ([]core.Run, error) {
	return s.runs, nil
}

func TestProviderStartEmitsEventsAndStopsOnContextCancel(t *testing.T) {
	runs := []core.Run{{ID: "42", Provider: "azure", WorkflowName: "Build", Status: core.RunStatusInProgress}}
	client := &stubAzureClient{runs: runs}
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
