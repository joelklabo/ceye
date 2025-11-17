package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joelklabo/ceye/internal/alerting"
	"github.com/joelklabo/ceye/internal/storage"
)

// BuildAlertEngine creates an alert engine from configuration
func BuildAlertEngine(cfg *AlertingConfig, store *storage.Storage) (*alerting.Engine, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}

	engine := alerting.NewEngine(store)

	// Build channels
	channels := make(map[string]alerting.AlertChannel)
	for name, channelCfg := range cfg.Channels {
		channel, err := buildChannel(name, channelCfg)
		if err != nil {
			return nil, fmt.Errorf("build channel %s: %w", name, err)
		}
		channels[name] = channel
	}

	// Build rules
	for _, ruleCfg := range cfg.Rules {
		rule, err := buildRule(ruleCfg, channels)
		if err != nil {
			return nil, fmt.Errorf("build rule %s: %w", ruleCfg.Name, err)
		}
		engine.AddRule(rule)
	}

	return engine, nil
}

// buildRule creates an AlertRule from config
func buildRule(cfg AlertRuleConfig, channels map[string]alerting.AlertChannel) (*alerting.AlertRule, error) {
	// Parse cooldown duration
	cooldown := 5 * time.Minute // default
	if cfg.Cooldown != "" {
		d, err := time.ParseDuration(cfg.Cooldown)
		if err != nil {
			return nil, fmt.Errorf("parse cooldown: %w", err)
		}
		cooldown = d
	}

	// Build condition
	condition, err := buildCondition(cfg.Condition)
	if err != nil {
		return nil, fmt.Errorf("build condition: %w", err)
	}

	// Map channel names to channel objects
	ruleChannels := make([]alerting.AlertChannel, 0, len(cfg.Channels))
	for _, channelName := range cfg.Channels {
		channel, ok := channels[channelName]
		if !ok {
			return nil, fmt.Errorf("channel not found: %s", channelName)
		}
		ruleChannels = append(ruleChannels, channel)
	}

	return &alerting.AlertRule{
		Name:      cfg.Name,
		Condition: condition,
		Channels:  ruleChannels,
		Providers: cfg.Providers,
		Cooldown:  cooldown,
		Enabled:   cfg.Enabled,
	}, nil
}

// buildCondition creates an AlertCondition from config
func buildCondition(cfg AlertConditionConfig) (alerting.AlertCondition, error) {
	switch cfg.Type {
	case "workflow_failed":
		return &alerting.WorkflowFailedCondition{}, nil

	case "high_failure_rate":
		threshold := cfg.Threshold
		if threshold == 0 {
			threshold = 0.2 // default 20%
		}
		
		period := 1 * time.Hour // default
		if cfg.Period != "" {
			d, err := time.ParseDuration(cfg.Period)
			if err != nil {
				return nil, fmt.Errorf("parse period: %w", err)
			}
			period = d
		}

		return &alerting.HighFailureRateCondition{
			Threshold: threshold,
			Period:    period,
		}, nil

	case "duration_spike":
		threshold := cfg.Threshold
		if threshold == 0 {
			threshold = 2.0 // default 2x
		}
		return &alerting.DurationSpikeCondition{
			Threshold: threshold,
		}, nil

	default:
		return nil, fmt.Errorf("unknown condition type: %s", cfg.Type)
	}
}

// buildChannel creates an AlertChannel from config
func buildChannel(name string, cfg AlertChannelConfig) (alerting.AlertChannel, error) {
	switch cfg.Type {
	case "slack":
		webhookURL := expandEnv(cfg.WebhookURL)
		if webhookURL == "" {
			return nil, fmt.Errorf("slack channel %s: webhook_url is required", name)
		}
		return &alerting.SlackChannel{
			WebhookURL: webhookURL,
			Timeout:    30 * time.Second,
		}, nil

	case "webhook":
		webhookURL := expandEnv(cfg.WebhookURL)
		if webhookURL == "" {
			return nil, fmt.Errorf("webhook channel %s: webhook_url is required", name)
		}
		return &alerting.WebhookChannel{
			URL:     webhookURL,
			Timeout: 30 * time.Second,
		}, nil

	case "log":
		return &alerting.LogChannel{}, nil

	default:
		return nil, fmt.Errorf("unknown channel type: %s (supported: slack, webhook, log)", cfg.Type)
	}
}

// expandEnv expands environment variables in the format ${VAR} or $VAR
func expandEnv(s string) string {
	return os.Expand(s, os.Getenv)
}

// ValidateAlertingConfig validates alerting configuration
func ValidateAlertingConfig(cfg *AlertingConfig) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	// Validate channels exist
	if len(cfg.Channels) == 0 {
		return fmt.Errorf("alerting enabled but no channels configured")
	}

	// Validate rules
	for _, rule := range cfg.Rules {
		if rule.Name == "" {
			return fmt.Errorf("rule missing name")
		}

		if rule.Condition.Type == "" {
			return fmt.Errorf("rule %s: condition type is required", rule.Name)
		}

		if len(rule.Channels) == 0 {
			return fmt.Errorf("rule %s: at least one channel is required", rule.Name)
		}

		// Check channels exist
		for _, channelName := range rule.Channels {
			if _, ok := cfg.Channels[channelName]; !ok {
				return fmt.Errorf("rule %s: channel not found: %s", rule.Name, channelName)
			}
		}
	}

	return nil
}
