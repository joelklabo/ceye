package config

import (
	"github.com/spf13/viper"

	"github.com/joelklabo/ceye/internal/providers"
)

// Config is the root configuration structure.
type Config struct {
	Providers []providers.ProviderConfig `mapstructure:"providers"`
}

// Load reads the configuration file at path into Config.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
