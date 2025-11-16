package providers

import (
	"strings"
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestEventValidator_ValidEvent(t *testing.T) {
	validator := NewEventValidator()

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:        "run-1",
				Provider:  "test",
				Status:    core.RunStatusInProgress,
				UpdatedAt: time.Now(),
				Duration:  5 * time.Minute,
			},
		},
		Timestamp: time.Now(),
	}

	if err := validator.Validate(event); err != nil {
		t.Errorf("valid event failed validation: %v", err)
	}
}

func TestRequiredFieldsRule_MissingProvider(t *testing.T) {
	rule := &RequiredFieldsRule{}

	event := core.RunEvent{
		Provider: "", // Missing
		Runs:     []core.Run{},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for missing provider")
	}
	if !strings.Contains(err.Error(), "provider name is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRequiredFieldsRule_MissingRunID(t *testing.T) {
	rule := &RequiredFieldsRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "", // Missing
				Provider: "test",
				Status:   core.RunStatusInProgress,
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for missing run ID")
	}
	if !strings.Contains(err.Error(), "ID is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRequiredFieldsRule_MissingRunProvider(t *testing.T) {
	rule := &RequiredFieldsRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "run-1",
				Provider: "", // Missing
				Status:   core.RunStatusInProgress,
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for missing run provider")
	}
	if !strings.Contains(err.Error(), "Provider is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRequiredFieldsRule_MissingRunStatus(t *testing.T) {
	rule := &RequiredFieldsRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "run-1",
				Provider: "test",
				Status:   "", // Missing
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for missing run status")
	}
	if !strings.Contains(err.Error(), "Status is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTimestampRule_FutureTimestamp(t *testing.T) {
	rule := &TimestampRule{}

	futureTime := time.Now().Add(48 * time.Hour)

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:        "run-1",
				Provider:  "test",
				Status:    core.RunStatusInProgress,
				UpdatedAt: futureTime,
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for future timestamp")
	}
	if !strings.Contains(err.Error(), "in the future") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTimestampRule_PastTimestamp(t *testing.T) {
	rule := &TimestampRule{}

	pastTime := time.Now().Add(-400 * 24 * time.Hour)

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:        "run-1",
				Provider:  "test",
				Status:    core.RunStatusInProgress,
				UpdatedAt: pastTime,
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for timestamp too far in past")
	}
	if !strings.Contains(err.Error(), "too far in the past") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTimestampRule_ValidTimestamps(t *testing.T) {
	rule := &TimestampRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:        "run-1",
				Provider:  "test",
				Status:    core.RunStatusInProgress,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
				StartedAt: time.Now().Add(-2 * time.Hour),
			},
		},
	}

	if err := rule.Validate(event); err != nil {
		t.Errorf("valid timestamps failed validation: %v", err)
	}
}

func TestStatusRule_InvalidStatus(t *testing.T) {
	rule := &StatusRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "run-1",
				Provider: "test",
				Status:   "invalid_status",
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestStatusRule_ValidStatuses(t *testing.T) {
	rule := &StatusRule{}

	validStatuses := []core.RunStatus{
		core.RunStatusUnknown,
		core.RunStatusQueued,
		core.RunStatusInProgress,
		core.RunStatusCompleted,
		core.RunStatusFailed,
		core.RunStatusCancelled,
	}

	for _, status := range validStatuses {
		event := core.RunEvent{
			Provider: "test",
			Runs: []core.Run{
				{
					ID:       "run-1",
					Provider: "test",
					Status:   status,
				},
			},
		}

		if err := rule.Validate(event); err != nil {
			t.Errorf("valid status %q failed validation: %v", status, err)
		}
	}
}

func TestDurationRule_NegativeDuration(t *testing.T) {
	rule := &DurationRule{}

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "run-1",
				Provider: "test",
				Status:   core.RunStatusInProgress,
				Duration: -5 * time.Minute,
			},
		},
	}

	err := rule.Validate(event)
	if err == nil {
		t.Error("expected error for negative duration")
	}
	if !strings.Contains(err.Error(), "duration cannot be negative") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDurationRule_ValidDurations(t *testing.T) {
	rule := &DurationRule{}

	durations := []time.Duration{
		0,
		1 * time.Second,
		5 * time.Minute,
		2 * time.Hour,
	}

	for _, duration := range durations {
		event := core.RunEvent{
			Provider: "test",
			Runs: []core.Run{
				{
					ID:       "run-1",
					Provider: "test",
					Status:   core.RunStatusInProgress,
					Duration: duration,
				},
			},
		}

		if err := rule.Validate(event); err != nil {
			t.Errorf("valid duration %v failed validation: %v", duration, err)
		}
	}
}

func TestEventValidator_MultipleRuns(t *testing.T) {
	validator := NewEventValidator()

	event := core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:       "run-1",
				Provider: "test",
				Status:   core.RunStatusInProgress,
			},
			{
				ID:       "run-2",
				Provider: "test",
				Status:   core.RunStatusCompleted,
			},
			{
				ID:       "", // Invalid
				Provider: "test",
				Status:   core.RunStatusFailed,
			},
		},
	}

	err := validator.Validate(event)
	if err == nil {
		t.Error("expected error for invalid run in batch")
	}
	if !strings.Contains(err.Error(), "run[2]") {
		t.Errorf("error should reference run[2], got: %v", err)
	}
}
