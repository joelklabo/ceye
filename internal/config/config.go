package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/joelklabo/ceye/internal/providers"
)

// Config is the root configuration structure.
type Config struct {
	Providers []providers.ProviderConfig `mapstructure:"providers"`
	Alerting  *AlertingConfig            `mapstructure:"alerting"`
}

// AlertingConfig configures the alerting engine
type AlertingConfig struct {
	Enabled  bool                       `mapstructure:"enabled"`
	Rules    []AlertRuleConfig          `mapstructure:"rules"`
	Channels map[string]AlertChannelConfig `mapstructure:"channels"`
}

// AlertRuleConfig defines an alert rule
type AlertRuleConfig struct {
	Name      string                 `mapstructure:"name"`
	Condition AlertConditionConfig   `mapstructure:"condition"`
	Channels  []string               `mapstructure:"channels"`
	Providers []string               `mapstructure:"providers"`
	Cooldown  string                 `mapstructure:"cooldown"` // Duration string like "30m"
	Enabled   bool                   `mapstructure:"enabled"`
}

// AlertConditionConfig defines when to trigger an alert
type AlertConditionConfig struct {
	Type      string                 `mapstructure:"type"` // workflow_failed, high_failure_rate, duration_spike, etc.
	Threshold float64                `mapstructure:"threshold,omitempty"`
	Period    string                 `mapstructure:"period,omitempty"` // Duration string
	Workflow  string                 `mapstructure:"workflow,omitempty"`
}

// AlertChannelConfig defines how to send alerts
type AlertChannelConfig struct {
	Type      string            `mapstructure:"type"` // slack, email, pagerduty, webhook
	WebhookURL string           `mapstructure:"webhook_url,omitempty"`
	APIKey    string            `mapstructure:"api_key,omitempty"`
	Email     EmailChannelConfig `mapstructure:"email,omitempty"`
}

// EmailChannelConfig configures email notifications
type EmailChannelConfig struct {
	Host     string   `mapstructure:"host"`
	Port     int      `mapstructure:"port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
}

var defaultSearchPaths = []string{
	".",
	filepath.Join(os.Getenv("HOME"), ".config", "ceye"),
}

// Load reads the configuration file at path into Config. If path is empty,
// Load searches default locations (current directory and ~/.config/ceye) for a
// file named ceye.(yaml|yml|json|toml).
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("ceye")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("ceye")

		for _, p := range defaultSearchPaths {
			if p == "" {
				continue
			}
			v.AddConfigPath(p)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
