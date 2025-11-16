package providers

import (
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

// EventValidator validates RunEvents before they're processed
type EventValidator struct {
	rules []ValidationRule
}

// ValidationRule is a single validation check
type ValidationRule interface {
	Validate(event core.RunEvent) error
}

// NewEventValidator creates a validator with standard rules
func NewEventValidator() *EventValidator {
	return &EventValidator{
		rules: []ValidationRule{
			&RequiredFieldsRule{},
			&TimestampRule{},
			&StatusRule{},
			&DurationRule{},
		},
	}
}

// Validate runs all validation rules on an event
func (v *EventValidator) Validate(event core.RunEvent) error {
	for _, rule := range v.rules {
		if err := rule.Validate(event); err != nil {
			return err
		}
	}
	return nil
}

// RequiredFieldsRule ensures critical fields are set
type RequiredFieldsRule struct{}

func (r *RequiredFieldsRule) Validate(event core.RunEvent) error {
	if event.Provider == "" {
		return fmt.Errorf("provider name is required")
	}

	for i, run := range event.Runs {
		if run.ID == "" {
			return fmt.Errorf("run[%d]: ID is required", i)
		}
		if run.Provider == "" {
			return fmt.Errorf("run[%d]: Provider is required", i)
		}
		if run.Status == "" {
			return fmt.Errorf("run[%d]: Status is required", i)
		}
	}

	return nil
}

// TimestampRule validates timestamps are reasonable
type TimestampRule struct{}

func (r *TimestampRule) Validate(event core.RunEvent) error {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-365 * 24 * time.Hour)

	for i, run := range event.Runs {
		if !run.UpdatedAt.IsZero() {
			if run.UpdatedAt.After(future) {
				return fmt.Errorf("run[%d]: UpdatedAt is in the future", i)
			}
			if run.UpdatedAt.Before(past) {
				return fmt.Errorf("run[%d]: UpdatedAt is too far in the past", i)
			}
		}

		if !run.StartedAt.IsZero() {
			if run.StartedAt.After(future) {
				return fmt.Errorf("run[%d]: StartedAt is in the future", i)
			}
			if run.StartedAt.Before(past) {
				return fmt.Errorf("run[%d]: StartedAt is too far in the past", i)
			}
		}
	}

	return nil
}

// StatusRule validates status values are valid
type StatusRule struct{}

func (r *StatusRule) Validate(event core.RunEvent) error {
	validStatuses := map[core.RunStatus]bool{
		core.RunStatusUnknown:    true,
		core.RunStatusQueued:     true,
		core.RunStatusInProgress: true,
		core.RunStatusCompleted:  true,
		core.RunStatusFailed:     true,
		core.RunStatusCancelled:  true,
	}

	for i, run := range event.Runs {
		if !validStatuses[run.Status] {
			return fmt.Errorf("run[%d]: invalid status %q", i, run.Status)
		}
	}

	return nil
}

// DurationRule validates durations are non-negative
type DurationRule struct{}

func (r *DurationRule) Validate(event core.RunEvent) error {
	for i, run := range event.Runs {
		if run.Duration < 0 {
			return fmt.Errorf("run[%d]: duration cannot be negative", i)
		}
	}
	return nil
}
