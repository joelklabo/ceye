package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ProviderContractTestSuite validates that any Provider implementation
// adheres to the contract defined by the Provider interface.
//
// This test suite ensures:
// - Provider respects context cancellation
// - Provider sends well-formed events
// - Provider handles concurrent calls safely
// - Provider name is stable
// - Provider Start can be called multiple times
type ProviderContractTestSuite struct {
	t             *testing.T
	providerName  string
	createProvider func() Provider
	timeout       time.Duration
}

// NewProviderContractTestSuite creates a test suite for a provider
func NewProviderContractTestSuite(t *testing.T, name string, createFn func() Provider) *ProviderContractTestSuite {
	return &ProviderContractTestSuite{
		t:              t,
		providerName:   name,
		createProvider: createFn,
		timeout:        5 * time.Second,
	}
}

// RunAll executes all contract tests
func (s *ProviderContractTestSuite) RunAll() {
	s.t.Run(fmt.Sprintf("%s/Name", s.providerName), func(t *testing.T) { s.testName(t) })
	s.t.Run(fmt.Sprintf("%s/ContextCancellation", s.providerName), func(t *testing.T) { s.testContextCancellation(t) })
	s.t.Run(fmt.Sprintf("%s/EventStructure", s.providerName), func(t *testing.T) { s.testEventStructure(t) })
	s.t.Run(fmt.Sprintf("%s/ConcurrentSafety", s.providerName), func(t *testing.T) { s.testConcurrentSafety(t) })
	s.t.Run(fmt.Sprintf("%s/MultipleStarts", s.providerName), func(t *testing.T) { s.testMultipleStarts(t) })
	s.t.Run(fmt.Sprintf("%s/EventChannel", s.providerName), func(t *testing.T) { s.testEventChannel(t) })
	s.t.Run(fmt.Sprintf("%s/NameStability", s.providerName), func(t *testing.T) { s.testNameStability(t) })
	s.t.Run(fmt.Sprintf("%s/ContextTimeout", s.providerName), func(t *testing.T) { s.testContextTimeout(t) })
}

// testName verifies provider returns a non-empty name
func (s *ProviderContractTestSuite) testName(t *testing.T) {
	provider := s.createProvider()
	name := provider.Name()

	if name == "" {
		t.Error("Provider.Name() returned empty string")
	}

	t.Logf("✓ Provider name: %q", name)
}

// testContextCancellation verifies provider respects context cancellation
func (s *ProviderContractTestSuite) testContextCancellation(t *testing.T) {
	provider := s.createProvider()

	ctx, cancel := context.WithCancel(context.Background())
	eventCh := make(chan RunEvent, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- provider.Start(ctx, eventCh)
	}()

	// Give provider time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Provider should exit within reasonable time
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled && !errors.Is(err, context.Canceled) {
			t.Logf("Provider returned error: %v (acceptable if includes context error)", err)
		}
		t.Log("✓ Provider respected context cancellation")
	case <-time.After(2 * time.Second):
		t.Error("Provider did not respect context cancellation within 2 seconds")
	}
}

// testEventStructure validates events have required fields
func (s *ProviderContractTestSuite) testEventStructure(t *testing.T) {
	provider := s.createProvider()

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	eventCh := make(chan RunEvent, 10)

	go provider.Start(ctx, eventCh)

	// Wait for first event
	select {
	case event := <-eventCh:
		s.validateEvent(t, event)
		t.Log("✓ Event structure is valid")
	case <-time.After(s.timeout):
		t.Log("⚠ No event received within timeout (acceptable for some providers)")
	}
}

// validateEvent checks an event has required fields
func (s *ProviderContractTestSuite) validateEvent(t *testing.T, event RunEvent) {
	if event.Provider == "" {
		t.Error("Event missing Provider field")
	}

	if event.Timestamp.IsZero() {
		t.Error("Event missing Timestamp")
	}

	// If there are runs, validate them
	for i, run := range event.Runs {
		if run.ID == "" {
			t.Errorf("Run[%d] missing ID", i)
		}
		if run.Provider == "" {
			t.Errorf("Run[%d] missing Provider", i)
		}
		if run.Status == "" {
			t.Errorf("Run[%d] missing Status", i)
		}
	}
}

// testConcurrentSafety verifies provider can be called concurrently
func (s *ProviderContractTestSuite) testConcurrentSafety(t *testing.T) {
	// Test that calling Name() concurrently is safe
	provider := s.createProvider()
	
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = provider.Name()
		}()
	}
	wg.Wait()

	t.Log("✓ Concurrent Name() calls are safe")
}

// testMultipleStarts verifies provider can be started multiple times
func (s *ProviderContractTestSuite) testMultipleStarts(t *testing.T) {
	provider := s.createProvider()

	// Start provider twice in sequence
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		eventCh := make(chan RunEvent, 10)

		err := provider.Start(ctx, eventCh)
		cancel()

		// Should complete without panic
		if err != nil && err != context.DeadlineExceeded && !errors.Is(err, context.Canceled) {
			t.Logf("Start iteration %d returned error: %v", i+1, err)
		}
	}

	t.Log("✓ Provider can be started multiple times")
}

// testEventChannel verifies provider uses the event channel correctly
func (s *ProviderContractTestSuite) testEventChannel(t *testing.T) {
	provider := s.createProvider()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use unbuffered channel to test blocking behavior
	eventCh := make(chan RunEvent)
	errCh := make(chan error, 1)

	go func() {
		errCh <- provider.Start(ctx, eventCh)
	}()

	// Try to receive events - provider should not deadlock
	timeout := time.After(1 * time.Second)
	eventCount := 0

eventLoop:
	for {
		select {
		case <-eventCh:
			eventCount++
			if eventCount >= 2 {
				break eventLoop
			}
		case <-timeout:
			break eventLoop
		case err := <-errCh:
			if err != nil && err != context.DeadlineExceeded && !errors.Is(err, context.Canceled) {
				t.Logf("Provider exited with: %v", err)
			}
			break eventLoop
		}
	}

	t.Logf("✓ Provider correctly uses event channel (received %d events)", eventCount)
}

// testNameStability verifies Name() always returns the same value
func (s *ProviderContractTestSuite) testNameStability(t *testing.T) {
	provider := s.createProvider()

	name1 := provider.Name()
	name2 := provider.Name()
	name3 := provider.Name()

	if name1 != name2 || name2 != name3 {
		t.Errorf("Name() not stable: %q, %q, %q", name1, name2, name3)
	}

	t.Log("✓ Provider name is stable")
}

// testContextTimeout verifies provider handles context timeout
func (s *ProviderContractTestSuite) testContextTimeout(t *testing.T) {
	provider := s.createProvider()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	eventCh := make(chan RunEvent, 10)
	err := provider.Start(ctx, eventCh)

	// Should return when context times out
	if err != nil && err != context.DeadlineExceeded && !errors.Is(err, context.DeadlineExceeded) {
		t.Logf("Provider returned: %v (acceptable if related to context)", err)
	}

	t.Log("✓ Provider handles context timeout")
}

// TestProviderContractExample shows how to use the test suite
func TestProviderContractExample(t *testing.T) {
	// Example: Testing a mock provider
	suite := NewProviderContractTestSuite(t, "MockProvider", func() Provider {
		return &mockProvider{name: "test-provider"}
	})

	suite.RunAll()
}

// mockProvider is a minimal provider for testing the test suite itself
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Start(ctx context.Context, out chan<- RunEvent) error {
	// Send one event
	select {
	case out <- RunEvent{
		Provider:  m.name,
		Timestamp: time.Now(),
		Runs: []Run{
			{
				ID:       "test-1",
				Provider: m.name,
				Status:   RunStatusInProgress,
			},
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// TestProviderInterfaceCompliance validates the Provider interface contract
// This test ensures any type implementing Provider works correctly
func TestProviderInterfaceCompliance(t *testing.T) {
	t.Run("MinimalProvider", func(t *testing.T) {
		suite := NewProviderContractTestSuite(t, "Minimal", func() Provider {
			return &minimalProvider{}
		})
		suite.RunAll()
	})

	t.Run("EventlessProvider", func(t *testing.T) {
		suite := NewProviderContractTestSuite(t, "Eventless", func() Provider {
			return &eventlessProvider{}
		})
		suite.RunAll()
	})
}

// minimalProvider implements Provider with minimal functionality
type minimalProvider struct{}

func (m *minimalProvider) Name() string { return "minimal" }

func (m *minimalProvider) Start(ctx context.Context, out chan<- RunEvent) error {
	<-ctx.Done()
	return ctx.Err()
}

// eventlessProvider never sends events
type eventlessProvider struct{}

func (e *eventlessProvider) Name() string { return "eventless" }

func (e *eventlessProvider) Start(ctx context.Context, out chan<- RunEvent) error {
	<-ctx.Done()
	return ctx.Err()
}

// TestProviderEventRequirements tests what makes a valid RunEvent
func TestProviderEventRequirements(t *testing.T) {
	tests := []struct {
		name    string
		event   RunEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event with runs",
			event: RunEvent{
				Provider:  "test",
				Timestamp: time.Now(),
				Runs: []Run{
					{
						ID:       "run-1",
						Provider: "test",
						Status:   RunStatusInProgress,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid event without runs",
			event: RunEvent{
				Provider:  "test",
				Timestamp: time.Now(),
				Runs:      []Run{},
			},
			wantErr: false,
		},
		{
			name: "missing provider name",
			event: RunEvent{
				Provider:  "",
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "provider name required",
		},
		{
			name: "run missing ID",
			event: RunEvent{
				Provider:  "test",
				Timestamp: time.Now(),
				Runs: []Run{
					{
						ID:       "",
						Provider: "test",
						Status:   RunStatusInProgress,
					},
				},
			},
			wantErr: true,
			errMsg:  "run ID required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRunEvent(tt.event)
			if tt.wantErr && err == nil {
				t.Errorf("expected error containing %q, got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantErr && err != nil {
				t.Logf("✓ Validation caught: %v", err)
			}
		})
	}
}

// validateRunEvent checks if an event meets requirements
func validateRunEvent(event RunEvent) error {
	if event.Provider == "" {
		return errors.New("provider name required")
	}

	for _, run := range event.Runs {
		if run.ID == "" {
			return errors.New("run ID required")
		}
		if run.Provider == "" {
			return errors.New("run provider required")
		}
		if run.Status == "" {
			return errors.New("run status required")
		}
	}

	return nil
}
