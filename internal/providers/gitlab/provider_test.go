package gitlab

import (
	"context"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestGitLabProviderEmitsRuns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	out := make(chan core.RunEvent, 1)
	provider := NewProvider(Config{Project: "example/proj"})
	go provider.Start(ctx, out)
	select {
	case event := <-out:
		if event.Provider != "gitlab" {
			t.Fatalf("expected provider gitlab, got %s", event.Provider)
		}
		if len(event.Runs) == 0 {
			t.Fatalf("expected runs, got none")
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("expected event from gitlab provider")
	}
}
