package manager

import (
	"path/filepath"
	"testing"

	"github.com/joelklabo/ceye/internal/providers"
)

func TestStoreAddEnableDisableRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	store, err := New(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	cfg := providers.ProviderConfig{Type: "demo", Runs: 3}
	record, err := store.Add(cfg)
	if err != nil {
		t.Fatalf("add provider: %v", err)
	}
	if !record.Enabled {
		t.Fatalf("expected record enabled by default")
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}

	if err := store.SetEnabled(record.ID, false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	updated, ok := store.Find(record.ID)
	if !ok || updated.Enabled {
		t.Fatalf("expected record disabled, got %+v", updated)
	}

	if err := store.Update(record.ID, providers.ProviderConfig{Type: "demo", Runs: 5}); err != nil {
		t.Fatalf("update: %v", err)
	}
	updated, _ = store.Find(record.ID)
	if updated.Config.Runs != 5 {
		t.Fatalf("expected runs updated, got %d", updated.Config.Runs)
	}

	if err := store.Remove(record.ID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(store.List()) != 0 {
		t.Fatalf("expected empty store")
	}
}

func TestStorePersistsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	store, err := New(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	cfg := providers.ProviderConfig{Type: "demo", Runs: 1}
	record, err := store.Add(cfg)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	// ensure load reuses entries
	reloaded, err := New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	found, ok := reloaded.Find(record.ID)
	if !ok {
		t.Fatalf("expected persisted record")
	}
	if found.Config.Runs != cfg.Runs {
		t.Fatalf("unexpected runs: %d", found.Config.Runs)
	}

	// Save a disabled entry and ensure EnabledRecords filters
	if err := reloaded.SetEnabled(record.ID, false); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	if len(reloaded.EnabledRecords()) != 0 {
		t.Fatalf("expected no enabled entries")
	}
}
