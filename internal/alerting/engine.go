package alerting

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/storage"
)

// Engine evaluates alert rules and dispatches notifications
type Engine struct {
	rules   []*AlertRule
	storage *storage.Storage
	store   *core.Store // For recording alert history
	state   map[string]*ruleState // rule name -> state
	mu      sync.RWMutex
}

// NewEngine creates a new alert engine
func NewEngine(storage *storage.Storage) *Engine {
	return &Engine{
		rules:   make([]*AlertRule, 0),
		storage: storage,
		store:   nil, // Set later via SetStore
		state:   make(map[string]*ruleState),
	}
}

// SetStore sets the core store for alert history
func (e *Engine) SetStore(store *core.Store) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store = store
}

// AddRule adds an alert rule to the engine
func (e *Engine) AddRule(rule *AlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, rule)
	if _, exists := e.state[rule.Name]; !exists {
		e.state[rule.Name] = &ruleState{}
	}
}

// RemoveRule removes an alert rule by name
func (e *Engine) RemoveRule(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, rule := range e.rules {
		if rule.Name == name {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			delete(e.state, name)
			return
		}
	}
}

// GetRules returns all configured rules
func (e *Engine) GetRules() []*AlertRule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*AlertRule, len(e.rules))
	copy(rules, e.rules)
	return rules
}

// Start begins processing run events and evaluating alert rules
func (e *Engine) Start(ctx context.Context, events <-chan core.RunEvent) {
	log.Printf("alerting: engine started")

	for {
		select {
		case <-ctx.Done():
			log.Printf("alerting: engine stopped")
			return

		case event := <-events:
			if event.Err != nil {
				continue
			}

			// Evaluate rules for each run in the event
			for _, run := range event.Runs {
				e.evaluateRun(run)
			}
		}
	}
}

// evaluateRun checks all rules against a run and fires alerts
func (e *Engine) evaluateRun(run core.Run) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()

	for _, rule := range e.rules {
		// Skip disabled rules
		if !rule.Enabled {
			continue
		}

		// Skip if provider filter doesn't match
		if len(rule.Providers) > 0 && !contains(rule.Providers, run.Provider) {
			continue
		}

		state := e.state[rule.Name]

		// Check cooldown
		if !state.lastAlertTime.IsZero() && now.Sub(state.lastAlertTime) < rule.Cooldown {
			continue
		}

		// Evaluate condition
		shouldAlert, message := rule.Condition.Check(run, e.storage)
		if !shouldAlert {
			continue
		}

		// Fire alert
		alert := Alert{
			RuleName:    rule.Name,
			Condition:   rule.Condition.Name(),
			Message:     message,
			Run:         run,
			TriggeredAt: now,
			Severity:    e.determineSeverity(rule),
		}

		// Record to store history
		if e.store != nil {
			e.store.RecordAlert(core.AlertRecord{
				RuleName:    alert.RuleName,
				Condition:   alert.Condition,
				Message:     alert.Message,
				Severity:    string(alert.Severity),
				Run:         alert.Run,
				TriggeredAt: alert.TriggeredAt,
			})
		}

		// Send to all channels
		for _, channel := range rule.Channels {
			go func(ch AlertChannel, a Alert) {
				if err := ch.Send(a); err != nil {
					log.Printf("alerting: failed to send alert via %s: %v", ch.Name(), err)
				} else {
					log.Printf("alerting: sent alert '%s' via %s", a.RuleName, ch.Name())
				}
			}(channel, alert)
		}

		// Update state
		state.lastAlertTime = now
		state.alertCount++
	}
}

// determineSeverity assigns severity based on rule type
func (e *Engine) determineSeverity(rule *AlertRule) AlertSeverity {
	// Simple heuristic based on condition type
	switch rule.Condition.(type) {
	case *WorkflowFailedCondition:
		return SeverityCritical
	case *HighFailureRateCondition:
		return SeverityCritical
	case *DurationSpikeCondition:
		return SeverityWarning
	case *BuildQueuedTooLongCondition:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
