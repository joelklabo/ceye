package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverConfigs(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "A"), 0o755)
	os.WriteFile(filepath.Join(root, "A", "ceye.yaml"), []byte("providers: []"), 0o644)
	os.WriteFile(filepath.Join(root, "ceye.json"), []byte("{}"), 0o644)

	matches, err := DiscoverConfigs(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(matches))
	}
}
