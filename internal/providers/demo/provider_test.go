package demo

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestDemoProviderEmitsRuns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan core.RunEvent, 1)
	provider := New(2, 20*time.Millisecond)

	go provider.Start(ctx, out)

	select {
	case event := <-out:
		if event.Provider != "demo" {
			t.Fatalf("expected provider demo, got %s", event.Provider)
		}
		if len(event.Runs) != 2 {
			t.Fatalf("expected 2 runs, got %d", len(event.Runs))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected event from demo provider")
	}
}
