package alerting

import (
	"context"
	"sync" // Added
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestEngineAddRemoveRules(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	rule := &AlertRule{
		Name:      "test-rule",
		Condition: &WorkflowFailedCondition{},
		Channels:  []AlertChannel{&LogChannel{}},
		Cooldown:  5 * time.Minute,
		Enabled:   true,
	}

	// Add rule
	engine.AddRule(rule)
	rules := engine.GetRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "test-rule" {
		t.Errorf("expected rule name 'test-rule', got %s", rules[0].Name)
	}

	// Remove rule
	engine.RemoveRule("test-rule")
	rules = engine.GetRules()
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules after removal, got %d", len(rules))
	}
}

func TestEngineCooldown(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	// Track alerts fired with synchronization
	alertCount := 0
	alertSent := make(chan struct{}, 10)
	testChannel := &testChannel{
		onSend: func(alert Alert) error {
			alertCount++
			alertSent <- struct{}{}
			return nil
		},
	}

	rule := &AlertRule{
		Name:      "cooldown-test",
		Condition: &WorkflowFailedCondition{},
		Channels:  []AlertChannel{testChannel},
		Cooldown:  100 * time.Millisecond, // Short cooldown for testing
		Enabled:   true,
	}

	engine.AddRule(rule)

	// Create failed run
	failedRun := core.Run{
		ID:           "1",
		Provider:     "test",
		Repo:         "test/repo",
		WorkflowName: "CI",
		Status:       core.RunStatusCompleted,
		Conclusion:   "failure",
		Branch:       "main",
	}

	// Fire first alert
	engine.evaluateRun(failedRun)
	<-alertSent // Wait for alert to be sent
	if alertCount != 1 {
		t.Errorf("expected 1 alert, got %d", alertCount)
	}

	// Try to fire immediately (should be blocked by cooldown)
	engine.evaluateRun(failedRun)
	select {
	case <-alertSent:
		t.Error("should not have sent alert during cooldown")
	case <-time.After(50 * time.Millisecond):
		// Expected - no alert sent
	}
	if alertCount != 1 {
		t.Errorf("expected still 1 alert due to cooldown, got %d", alertCount)
	}

	// Wait for cooldown to expire
	time.Sleep(100 * time.Millisecond)

	// Fire again (should work now)
	engine.evaluateRun(failedRun)
	<-alertSent // Wait for second alert
	if alertCount != 2 {
		t.Errorf("expected 2 alerts after cooldown, got %d", alertCount)
	}
}

func TestEngineProviderFiltering(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	alertCount := 0
	alertSent := make(chan struct{}, 10)
	testChannel := &testChannel{
		onSend: func(alert Alert) error {
			alertCount++
			alertSent <- struct{}{}
			return nil
		},
	}

	// Rule only for GitHub provider
	rule := &AlertRule{
		Name:      "github-only",
		Condition: &WorkflowFailedCondition{},
		Channels:  []AlertChannel{testChannel},
		Cooldown:  time.Millisecond,
		Enabled:   true,
		Providers: []string{"github"},
	}

	engine.AddRule(rule)

	// GitHub run (should alert)
	githubRun := core.Run{
		Provider:   "github",
		Status:     core.RunStatusCompleted,
		Conclusion: "failure",
	}
	engine.evaluateRun(githubRun)
	<-alertSent // Wait for alert
	if alertCount != 1 {
		t.Errorf("expected 1 alert for github run, got %d", alertCount)
	}

	// Azure run (should not alert due to filter)
	azureRun := core.Run{
		Provider:   "azure",
		Status:     core.RunStatusCompleted,
		Conclusion: "failure",
	}
	engine.evaluateRun(azureRun)
	select {
	case <-alertSent:
		t.Error("should not have sent alert for filtered provider")
	case <-time.After(50 * time.Millisecond):
		// Expected - no alert sent
	}
	if alertCount != 1 {
		t.Errorf("expected still 1 alert (azure filtered), got %d", alertCount)
	}
}

func TestEngineDisabledRule(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	alertCount := 0
	testChannel := &testChannel{
		onSend: func(alert Alert) error {
			alertCount++
			return nil
		},
	}

	rule := &AlertRule{
		Name:      "disabled-rule",
		Condition: &WorkflowFailedCondition{},
		Channels:  []AlertChannel{testChannel},
		Cooldown:  time.Millisecond,
		Enabled:   false, // Disabled
	}

	engine.AddRule(rule)

	failedRun := core.Run{
		Status:     core.RunStatusCompleted,
		Conclusion: "failure",
	}

	engine.evaluateRun(failedRun)
	if alertCount != 0 {
		t.Errorf("expected 0 alerts for disabled rule, got %d", alertCount)
	}
}

func TestEngineStart(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	var mu sync.Mutex // Added mutex
	alertCount := 0
	testChannel := &testChannel{
		onSend: func(alert Alert) error {
			mu.Lock() // Lock before write
			alertCount++
			mu.Unlock() // Unlock after write
			return nil
		},
	}

	rule := &AlertRule{
		Name:      "start-test",
		Condition: &WorkflowFailedCondition{},
		Channels:  []AlertChannel{testChannel},
		Cooldown:  time.Millisecond,
		Enabled:   true,
	}

	engine.AddRule(rule)

	// Create event channel
	events := make(chan core.RunEvent, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start engine in background
	go engine.Start(ctx, events)

	// Send failed run event
	events <- core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				Status:     core.RunStatusCompleted,
				Conclusion: "failure",
			},
		},
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock() // Lock before read
	if alertCount != 1 {
		t.Errorf("expected 1 alert from event, got %d", alertCount)
	}
	mu.Unlock() // Unlock after read

	// Send success event (should not alert)
	events <- core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				Status:     core.RunStatusCompleted,
				Conclusion: "success",
			},
		},
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock() // Lock before read
	if alertCount != 1 {
		t.Errorf("expected still 1 alert (success doesn't trigger), got %d", alertCount)
	}
	mu.Unlock() // Unlock after read
}

func TestEngineSeverityAssignment(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	engine := NewEngine(store)

	tests := []struct {
		name             string
		condition        AlertCondition
		expectedSeverity AlertSeverity
	}{
		{
			name:             "workflow failed",
			condition:        &WorkflowFailedCondition{},
			expectedSeverity: SeverityCritical,
		},
		{
			name:             "high failure rate",
			condition:        &HighFailureRateCondition{Threshold: 0.2, Period: 24 * time.Hour, MinRuns: 5},
			expectedSeverity: SeverityCritical,
		},
		{
			name:             "duration spike",
			condition:        &DurationSpikeCondition{Threshold: 2.0},
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "queued too long",
			condition:        &BuildQueuedTooLongCondition{MaxQueueTime: 10 * time.Minute},
			expectedSeverity: SeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &AlertRule{
				Name:      tt.name,
				Condition: tt.condition,
			}

			severity := engine.determineSeverity(rule)
			if severity != tt.expectedSeverity {
				t.Errorf("determineSeverity() = %v, want %v", severity, tt.expectedSeverity)
			}
		})
	}
}

// testChannel is a mock alert channel for testing
type testChannel struct {
	onSend func(alert Alert) error
}

func (c *testChannel) Name() string {
	return "test"
}

func (c *testChannel) Send(alert Alert) error {
	if c.onSend != nil {
		return c.onSend(alert)
	}
	return nil
}
