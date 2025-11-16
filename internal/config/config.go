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
