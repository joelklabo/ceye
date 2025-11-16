package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// DiscoverConfigs walks root and returns paths matching ceye.(yaml|yml|json|toml).
func DiscoverConfigs(root string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("root path is required")
	}
	var matches []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(filepath.Base(path))
		if name == "ceye.yaml" || name == "ceye.yml" || name == "ceye.json" || name == "ceye.toml" {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover configs: %w", err)
	}
	return matches, nil
}
