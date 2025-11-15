package providers

import (
	"fmt"
	"strings"
)

// DisplayName returns the friendly label for a provider config.
func DisplayName(cfg ProviderConfig) string {
	if name := strings.TrimSpace(cfg.DisplayName); name != "" {
		return name
	}
	if cfg.Type != "" {
		return fmt.Sprintf("%s provider", cfg.Type)
	}
	return "provider"
}
