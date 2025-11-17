package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// LogChannel logs alerts to stdout (useful for testing)
type LogChannel struct{}

func (c *LogChannel) Name() string {
	return "log"
}

func (c *LogChannel) Send(alert Alert) error {
	log.Printf("[ALERT %s] %s: %s (Run: %s/%s#%s)",
		alert.Severity,
		alert.RuleName,
		alert.Message,
		alert.Run.Repo,
		alert.Run.WorkflowName,
		alert.Run.ID)
	return nil
}

// WebhookChannel sends alerts to a webhook URL
type WebhookChannel struct {
	URL     string
	Timeout time.Duration
}

func (c *WebhookChannel) Name() string {
	return fmt.Sprintf("webhook(%s)", c.URL)
}

func (c *WebhookChannel) Send(alert Alert) error {
	payload := map[string]interface{}{
		"rule":        alert.RuleName,
		"condition":   alert.Condition,
		"message":     alert.Message,
		"severity":    alert.Severity,
		"triggered_at": alert.TriggeredAt.Format(time.RFC3339),
		"run": map[string]interface{}{
			"id":           alert.Run.ID,
			"provider":     alert.Run.Provider,
			"repo":         alert.Run.Repo,
			"workflow":     alert.Run.WorkflowName,
			"status":       alert.Run.Status,
			"conclusion":   alert.Run.Conclusion,
			"branch":       alert.Run.Branch,
			"commit_sha":   alert.Run.CommitSHA,
			"url":          alert.Run.URL,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	client := &http.Client{
		Timeout: c.Timeout,
	}

	resp, err := client.Post(c.URL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackChannel sends alerts to Slack via webhook
type SlackChannel struct {
	WebhookURL string
	Timeout    time.Duration
}

func (c *SlackChannel) Name() string {
	return "slack"
}

func (c *SlackChannel) Send(alert Alert) error {
	// Format message for Slack
	color := "good"
	emoji := ":white_check_mark:"
	
	switch alert.Severity {
	case SeverityWarning:
		color = "warning"
		emoji = ":warning:"
	case SeverityCritical:
		color = "danger"
		emoji = ":rotating_light:"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      fmt.Sprintf("%s %s", emoji, alert.RuleName),
				"text":       alert.Message,
				"footer":     fmt.Sprintf("ceye | %s", alert.Run.Provider),
				"ts":         alert.TriggeredAt.Unix(),
				"mrkdwn_in":  []string{"text"},
				"fields": []map[string]interface{}{
					{
						"title": "Workflow",
						"value": alert.Run.WorkflowName,
						"short": true,
					},
					{
						"title": "Repository",
						"value": alert.Run.Repo,
						"short": true,
					},
					{
						"title": "Branch",
						"value": alert.Run.Branch,
						"short": true,
					},
					{
						"title": "Status",
						"value": fmt.Sprintf("%s (%s)", alert.Run.Status, alert.Run.Conclusion),
						"short": true,
					},
				},
				"actions": []map[string]interface{}{
					{
						"type": "button",
						"text": "View Run",
						"url":  alert.Run.URL,
					},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	client := &http.Client{
		Timeout: c.Timeout,
	}

	resp, err := client.Post(c.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}
