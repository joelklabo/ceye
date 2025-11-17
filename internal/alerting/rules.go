package alerting

import (
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/storage"
)

// AlertRule defines when and how to send alerts
type AlertRule struct {
	Name        string
	Description string
	Condition   AlertCondition
	Channels    []AlertChannel
	Cooldown    time.Duration // Minimum time between alerts for same rule
	Enabled     bool
	Providers   []string // Empty = all providers
}

// AlertCondition determines if an alert should fire
type AlertCondition interface {
	// Check evaluates if the condition is met
	// Returns true if alert should fire, along with context message
	Check(run core.Run, storage *storage.Storage) (bool, string)
	
	// Name returns a human-readable name for the condition
	Name() string
}

// AlertChannel sends alert notifications
type AlertChannel interface {
	// Send delivers an alert notification
	Send(alert Alert) error
	
	// Name returns the channel name
	Name() string
}

// Alert represents a triggered alert with context
type Alert struct {
	RuleName    string
	Condition   string
	Message     string
	Run         core.Run
	TriggeredAt time.Time
	Severity    AlertSeverity
}

// AlertSeverity indicates alert importance
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// RuleState tracks per-rule alerting state
type ruleState struct {
	lastAlertTime time.Time
	alertCount    int
}
