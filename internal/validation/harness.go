package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
	"github.com/joelklabo/ceye/internal/providers/github"
)

// Discrepancy represents a mismatch between webhook and polling data
type Discrepancy struct {
	Timestamp    time.Time   `json:"timestamp"`
	Type         string      `json:"type"` // "missing_in_webhook", "missing_in_polling", "data_mismatch"
	RunID        string      `json:"run_id"`
	WebhookData  *core.Run   `json:"webhook_data,omitempty"`
	PollingData  *core.Run   `json:"polling_data,omitempty"`
	TimeDelta    time.Duration `json:"time_delta,omitempty"`
}

// ValidationMetrics tracks performance and accuracy metrics
type ValidationMetrics struct {
	Timestamp           time.Time      `json:"timestamp"`
	WebhookRunCount     int            `json:"webhook_run_count"`
	PollingRunCount     int            `json:"polling_run_count"`
	MatchedRuns         int            `json:"matched_runs"`
	MissingInWebhook    int            `json:"missing_in_webhook"`
	MissingInPolling    int            `json:"missing_in_polling"`
	AvgWebhookLatency   time.Duration  `json:"avg_webhook_latency"`
	AvgPollingLatency   time.Duration  `json:"avg_polling_latency"`
	WebhookAdvantage    time.Duration  `json:"webhook_advantage"` // How much faster webhooks are
}

// Harness runs dual-mode validation: webhook + polling simultaneously
type Harness struct {
	repos           []github.RepoConfig
	webhookStore    *core.Store
	pollingStore    *core.Store
	discrepancies   []Discrepancy
	mu              sync.RWMutex
	
	// Timing tracking
	webhookTimes    map[string]time.Time // runID -> arrival time
	pollingTimes    map[string]time.Time // runID -> discovery time
	timesMu         sync.RWMutex
	
	metricsLog      *os.File
	compareInterval time.Duration
}

// NewHarness creates a validation harness for the given repositories
func NewHarness(repos []github.RepoConfig) (*Harness, error) {
	// Ensure tmp/ directory exists
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return nil, fmt.Errorf("failed to create tmp directory: %w", err)
	}
	
	// Create metrics log file in tmp/ directory
	metricsPath := "tmp/validation-metrics.jsonl"
	metricsFile, err := os.OpenFile(metricsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open metrics log: %w", err)
	}

	return &Harness{
		repos:           repos,
		webhookStore:    core.NewStore(),
		pollingStore:    core.NewStore(),
		discrepancies:   []Discrepancy{},
		webhookTimes:    make(map[string]time.Time),
		pollingTimes:    make(map[string]time.Time),
		metricsLog:      metricsFile,
		compareInterval: 30 * time.Second,
	}, nil
}

// WebhookStore returns the webhook-mode store
func (h *Harness) WebhookStore() *core.Store {
	return h.webhookStore
}

// PollingStore returns the polling-mode store
func (h *Harness) PollingStore() *core.Store {
	return h.pollingStore
}

// Start begins dual-mode monitoring
func (h *Harness) Start(ctx context.Context) error {
	// Create GitHub client
	client := github.NewCLIClient()
	
	// Create webhook provider
	webhookProvider := github.NewProvider(client, h.repos)
	webhookProvider.SetWebhookMode(true)
	
	// Create polling provider  
	pollingProvider := github.NewProvider(client, h.repos)
	pollingProvider.SetWebhookMode(false)
	
	// Wrap both providers with SafeProvider
	safeWebhook := providers.NewSafeProvider(webhookProvider)
	safePolling := providers.NewSafeProvider(pollingProvider)
	
	// Create event channels
	webhookEvents := make(chan core.RunEvent, 100)
	pollingEvents := make(chan core.RunEvent, 100)
	
	// Start webhook provider
	go func() {
		if err := safeWebhook.Start(ctx, webhookEvents); err != nil && ctx.Err() == nil {
			log.Printf("Webhook provider error: %v", err)
		}
	}()
	
	// Start polling provider
	go func() {
		if err := safePolling.Start(ctx, pollingEvents); err != nil && ctx.Err() == nil {
			log.Printf("Polling provider error: %v", err)
		}
	}()
	
	// Process webhook events
	go h.processWebhookEvents(ctx, webhookEvents)
	
	// Process polling events
	go h.processPollingEvents(ctx, pollingEvents)
	
	// Periodically compare stores and log metrics
	go h.compareLoop(ctx)
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Close metrics log
	if h.metricsLog != nil {
		h.metricsLog.Close()
	}
	
	return ctx.Err()
}

func (h *Harness) processWebhookEvents(ctx context.Context, events chan core.RunEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-events:
			// Merge into webhook store
			h.webhookStore.Merge(event)
			
			// Record arrival times
			now := time.Now()
			for _, run := range event.Runs {
				h.RecordWebhookEvent(run, now)
			}
		}
	}
}

func (h *Harness) processPollingEvents(ctx context.Context, events chan core.RunEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-events:
			// Merge into polling store
			h.pollingStore.Merge(event)
			
			// Record discovery times
			now := time.Now()
			for _, run := range event.Runs {
				h.RecordPollingEvent(run, now)
			}
		}
	}
}

func (h *Harness) compareLoop(ctx context.Context) {
	ticker := time.NewTicker(h.compareInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			discrepancies := h.Compare()
			if len(discrepancies) > 0 {
				h.logDiscrepancies(discrepancies)
			}
			
			metrics := h.GetMetrics()
			h.logMetrics(metrics)
		}
	}
}

// Compare compares webhook and polling stores and returns discrepancies
func (h *Harness) Compare() []Discrepancy {
	webhookRuns := h.webhookStore.ListRuns("")
	pollingRuns := h.pollingStore.ListRuns("")
	
	// Create maps for efficient lookup
	webhookMap := make(map[string]core.Run)
	for _, run := range webhookRuns {
		webhookMap[run.ID] = run
	}
	
	pollingMap := make(map[string]core.Run)
	for _, run := range pollingRuns {
		pollingMap[run.ID] = run
	}
	
	var discrepancies []Discrepancy
	
	// Check for runs in polling but not in webhook
	for _, run := range pollingRuns {
		if _, exists := webhookMap[run.ID]; !exists {
			discrepancies = append(discrepancies, Discrepancy{
				Timestamp:   time.Now(),
				Type:        "missing_in_webhook",
				RunID:       run.ID,
				PollingData: &run,
			})
		}
	}
	
	// Check for runs in webhook but not in polling
	for _, run := range webhookRuns {
		if _, exists := pollingMap[run.ID]; !exists {
			discrepancies = append(discrepancies, Discrepancy{
				Timestamp:   time.Now(),
				Type:        "missing_in_polling",
				RunID:       run.ID,
				WebhookData: &run,
			})
		}
	}
	
	// Store discrepancies
	h.mu.Lock()
	h.discrepancies = append(h.discrepancies, discrepancies...)
	h.mu.Unlock()
	
	return discrepancies
}

// RecordWebhookEvent records when a run arrived via webhook
func (h *Harness) RecordWebhookEvent(run core.Run, arrivalTime time.Time) {
	h.timesMu.Lock()
	defer h.timesMu.Unlock()
	h.webhookTimes[run.ID] = arrivalTime
}

// RecordPollingEvent records when a run was discovered via polling
func (h *Harness) RecordPollingEvent(run core.Run, discoveryTime time.Time) {
	h.timesMu.Lock()
	defer h.timesMu.Unlock()
	h.pollingTimes[run.ID] = discoveryTime
}

// GetMetrics calculates current validation metrics
func (h *Harness) GetMetrics() ValidationMetrics {
	webhookRuns := h.webhookStore.ListRuns("")
	pollingRuns := h.pollingStore.ListRuns("")
	
	// Create ID sets
	webhookIDs := make(map[string]bool)
	for _, run := range webhookRuns {
		webhookIDs[run.ID] = true
	}
	
	pollingIDs := make(map[string]bool)
	for _, run := range pollingRuns {
		pollingIDs[run.ID] = true
	}
	
	// Count matches and mismatches
	matched := 0
	for id := range webhookIDs {
		if pollingIDs[id] {
			matched++
		}
	}
	
	missingInWebhook := len(pollingIDs) - matched
	missingInPolling := len(webhookIDs) - matched
	
	// Calculate average latencies and advantages
	h.timesMu.RLock()
	defer h.timesMu.RUnlock()
	
	var totalAdvantage time.Duration
	advantageCount := 0
	
	for id := range webhookIDs {
		if webhookTime, hasWebhook := h.webhookTimes[id]; hasWebhook {
			if pollingTime, hasPolling := h.pollingTimes[id]; hasPolling {
				// Calculate advantage (how much faster webhook was)
				advantage := pollingTime.Sub(webhookTime)
				if advantage > 0 {
					totalAdvantage += advantage
					advantageCount++
				}
			}
		}
	}
	
	var avgAdvantage time.Duration
	if advantageCount > 0 {
		avgAdvantage = totalAdvantage / time.Duration(advantageCount)
	}
	
	return ValidationMetrics{
		Timestamp:         time.Now(),
		WebhookRunCount:   len(webhookRuns),
		PollingRunCount:   len(pollingRuns),
		MatchedRuns:       matched,
		MissingInWebhook:  missingInWebhook,
		MissingInPolling:  missingInPolling,
		AvgWebhookLatency: 0, // TODO: Calculate from actual webhook arrival times
		AvgPollingLatency: 0, // TODO: Calculate from actual polling discovery times
		WebhookAdvantage:  avgAdvantage,
	}
}

func (h *Harness) logDiscrepancies(discrepancies []Discrepancy) {
	for _, d := range discrepancies {
		log.Printf("⚠️  Discrepancy detected: %s - RunID: %s", d.Type, d.RunID)
	}
}

func (h *Harness) logMetrics(metrics ValidationMetrics) {
	// Log to stdout
	log.Printf("📊 Validation Metrics: webhook=%d polling=%d matched=%d advantage=%v",
		metrics.WebhookRunCount,
		metrics.PollingRunCount,
		metrics.MatchedRuns,
		metrics.WebhookAdvantage)
	
	// Log to file as JSON
	if h.metricsLog != nil {
		json.NewEncoder(h.metricsLog).Encode(metrics)
	}
}

// GetDiscrepancies returns all recorded discrepancies
func (h *Harness) GetDiscrepancies() []Discrepancy {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	result := make([]Discrepancy, len(h.discrepancies))
	copy(result, h.discrepancies)
	return result
}
