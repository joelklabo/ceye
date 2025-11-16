package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// Mock providers for testing

type panicProvider struct {
	name string
}

func (p *panicProvider) Name() string { return p.name }

func (p *panicProvider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	panic("intentional panic for testing")
}

type invalidEventProvider struct {
	name string
}

func (p *invalidEventProvider) Name() string { return p.name }

func (p *invalidEventProvider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	// Send invalid event (missing required fields)
	select {
	case out <- core.RunEvent{
		Provider: p.name,
		Runs: []core.Run{
			{ID: ""}, // Invalid: missing ID
		},
		Timestamp: time.Now(),
	}:
	case <-ctx.Done():
		return ctx.Err()
	}
	<-ctx.Done()
	return ctx.Err()
}

type goodProvider struct {
	name string
}

func (p *goodProvider) Name() string { return p.name }

func (p *goodProvider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	select {
	case out <- core.RunEvent{
		Provider: p.name,
		Runs: []core.Run{
			{
				ID:       "test-run-1",
				Provider: p.name,
				Status:   core.RunStatusInProgress,
			},
		},
		Timestamp: time.Now(),
	}:
	case <-ctx.Done():
		return ctx.Err()
	}
	<-ctx.Done()
	return ctx.Err()
}

func TestSafeProvider_PanicRecovery(t *testing.T) {
	inner := &panicProvider{name: "panic-test"}
	safe := NewSafeProvider(inner)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	eventCh := make(chan core.RunEvent, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- safe.Start(ctx, eventCh)
	}()

	// Should receive error event about the panic
	select {
	case event := <-eventCh:
		if event.Err == nil {
			t.Error("expected error event, got nil")
		}
		if event.Provider != "panic-test" {
			t.Errorf("provider = %q, want panic-test", event.Provider)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for panic error event")
	}

	// Should return error
	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected error return, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error return")
	}
}

func TestSafeProvider_InvalidEvent(t *testing.T) {
	inner := &invalidEventProvider{name: "invalid-test"}
	safe := NewSafeProvider(inner)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	eventCh := make(chan core.RunEvent, 10)

	go safe.Start(ctx, eventCh)

	// Should receive validation error event
	select {
	case event := <-eventCh:
		if event.Err == nil {
			t.Fatal("expected validation error event, got nil")
		}
		if event.Provider != "invalid-test" {
			t.Errorf("provider = %q, want invalid-test", event.Provider)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for validation error event")
	}
}

func TestSafeProvider_ValidEvent(t *testing.T) {
	inner := &goodProvider{name: "good-test"}
	safe := NewSafeProvider(inner)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	eventCh := make(chan core.RunEvent, 10)

	go safe.Start(ctx, eventCh)

	// Should receive valid event
	select {
	case event := <-eventCh:
		if event.Err != nil {
			t.Errorf("expected valid event, got error: %v", event.Err)
		}
		if event.Provider != "good-test" {
			t.Errorf("provider = %q, want good-test", event.Provider)
		}
		if len(event.Runs) != 1 {
			t.Fatalf("len(Runs) = %d, want 1", len(event.Runs))
		}
		if event.Runs[0].ID != "test-run-1" {
			t.Errorf("run ID = %q, want test-run-1", event.Runs[0].ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for valid event")
	}
}

func TestSafeProvider_Name(t *testing.T) {
	inner := &goodProvider{name: "test-provider"}
	safe := NewSafeProvider(inner)

	if safe.Name() != "test-provider" {
		t.Errorf("Name() = %q, want test-provider", safe.Name())
	}
}

func TestSafeProvider_ContextCancellation(t *testing.T) {
	inner := &goodProvider{name: "cancel-test"}
	safe := NewSafeProvider(inner)

	ctx, cancel := context.WithCancel(context.Background())
	eventCh := make(chan core.RunEvent, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- safe.Start(ctx, eventCh)
	}()

	// Let it start
	time.Sleep(100 * time.Millisecond)

	// Cancel
	cancel()

	// Should return context error
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for context cancellation")
	}
}

func TestSafeProvider_Unwrap(t *testing.T) {
	inner := &goodProvider{name: "unwrap-test"}
	safe := NewSafeProvider(inner)

	unwrapped := safe.Unwrap()
	if unwrapped != inner {
		t.Error("Unwrap() did not return inner provider")
	}
}
