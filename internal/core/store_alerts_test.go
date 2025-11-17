package core

import (
	"testing"
	"time"
)

func TestStore_RecordAlert(t *testing.T) {
	store := NewStore()

	alert1 := AlertRecord{
		RuleName:    "test-rule",
		Condition:   "workflow_failed",
		Message:     "Build failed",
		Severity:    "critical",
		Run:         Run{ID: "1", Repo: "test/repo"},
		TriggeredAt: time.Now(),
	}

	alert2 := AlertRecord{
		RuleName:    "test-rule-2",
		Condition:   "high_failure_rate",
		Message:     "High failure rate detected",
		Severity:    "warning",
		Run:         Run{ID: "2", Repo: "test/repo"},
		TriggeredAt: time.Now(),
	}

	// Record alerts
	store.RecordAlert(alert1)
	store.RecordAlert(alert2)

	// Get recent alerts
	alerts := store.GetRecentAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}

	// Alerts should be in reverse chronological order (newest first)
	if alerts[0].RuleName != "test-rule-2" {
		t.Errorf("expected first alert to be 'test-rule-2', got %s", alerts[0].RuleName)
	}
	if alerts[1].RuleName != "test-rule" {
		t.Errorf("expected second alert to be 'test-rule', got %s", alerts[1].RuleName)
	}
}

func TestStore_GetRecentAlertsLimit(t *testing.T) {
	store := NewStore()

	// Add 10 alerts
	for i := 0; i < 10; i++ {
		store.RecordAlert(AlertRecord{
			RuleName:    "test-rule",
			Message:     "Test alert",
			Severity:    "info",
			TriggeredAt: time.Now(),
		})
	}

	// Get only 5
	alerts := store.GetRecentAlerts(5)
	if len(alerts) != 5 {
		t.Errorf("expected 5 alerts, got %d", len(alerts))
	}

	// Get all
	alerts = store.GetRecentAlerts(100)
	if len(alerts) != 10 {
		t.Errorf("expected 10 alerts, got %d", len(alerts))
	}

	// Get with limit 0 (should return all)
	alerts = store.GetRecentAlerts(0)
	if len(alerts) != 10 {
		t.Errorf("expected 10 alerts with limit 0, got %d", len(alerts))
	}
}

func TestStore_AlertHistoryLimit(t *testing.T) {
	store := NewStore()

	// Add 150 alerts (more than the 100 limit)
	for i := 0; i < 150; i++ {
		store.RecordAlert(AlertRecord{
			RuleName:    "test-rule",
			Message:     "Test alert",
			Severity:    "info",
			TriggeredAt: time.Now(),
		})
	}

	// Should only keep last 100
	alerts := store.GetRecentAlerts(0)
	if len(alerts) != 100 {
		t.Errorf("expected 100 alerts (limit), got %d", len(alerts))
	}

	// Check alert count
	count := store.GetAlertCount()
	if count != 100 {
		t.Errorf("expected alert count 100, got %d", count)
	}
}

func TestStore_GetAlertCount(t *testing.T) {
	store := NewStore()

	if count := store.GetAlertCount(); count != 0 {
		t.Errorf("expected 0 alerts initially, got %d", count)
	}

	// Add some alerts
	for i := 0; i < 5; i++ {
		store.RecordAlert(AlertRecord{
			RuleName: "test",
			Severity: "info",
			TriggeredAt: time.Now(),
		})
	}

	if count := store.GetAlertCount(); count != 5 {
		t.Errorf("expected 5 alerts, got %d", count)
	}
}

func TestStore_AlertsConcurrentAccess(t *testing.T) {
	store := NewStore()
	done := make(chan bool)

	// Write alerts concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				store.RecordAlert(AlertRecord{
					RuleName:    "test",
					Severity:    "info",
					TriggeredAt: time.Now(),
				})
			}
			done <- true
		}()
	}

	// Wait for all writers
	for i := 0; i < 10; i++ {
		<-done
	}

	// Read concurrently
	for i := 0; i < 10; i++ {
		go func() {
			_ = store.GetRecentAlerts(10)
			_ = store.GetAlertCount()
			done <- true
		}()
	}

	// Wait for all readers
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 100 alerts (due to limit)
	count := store.GetAlertCount()
	if count != 100 {
		t.Errorf("expected 100 alerts after concurrent writes, got %d", count)
	}
}
