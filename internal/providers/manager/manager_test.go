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
	record, err := store.Add("test", cfg)
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

	if err := store.SetEnabled("test", record.ID, false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	updated, ok := store.Find(record.ID)
	if !ok || updated.Enabled {
		t.Fatalf("expected record disabled, got %+v", updated)
	}

	if err := store.Update("test", record.ID, providers.ProviderConfig{Type: "demo", Runs: 5}); err != nil {
		t.Fatalf("update: %v", err)
	}
	updated, _ = store.Find(record.ID)
	if updated.Config.Runs != 5 {
		t.Fatalf("expected runs updated, got %d", updated.Config.Runs)
	}

	if err := store.Remove("test", record.ID); err != nil {
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
	record, err := store.Add("test", cfg)
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
	if err := reloaded.SetEnabled("test", record.ID, false); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	if len(reloaded.EnabledRecords()) != 0 {
		t.Fatalf("expected no enabled entries")
	}
}

func TestStoreAuditRecordsActions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	store, err := New(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	cfg := providers.ProviderConfig{Type: "demo", Runs: 1}
	record, err := store.Add("test", cfg)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if entries := store.Audit(); len(entries) != 1 || entries[0].Action != "add" || entries[0].Actor != "test" {
		t.Fatalf("unexpected audit log after add: %+v", store.Audit())
	}

	if err := store.SetEnabled("test", record.ID, false); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	if entries := store.Audit(); len(entries) == 0 || entries[0].Action != "disable" {
		t.Fatalf("expected disable audit entry, got %+v", entries)
	}

	if err := store.Update("test", record.ID, providers.ProviderConfig{Type: "demo", Runs: 2}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if entries := store.Audit(); len(entries) == 0 || entries[0].Action != "update" {
		t.Fatalf("expected update audit entry, got %+v", entries)
	}

	if err := store.Remove("test", record.ID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if entries := store.Audit(); len(entries) == 0 || entries[0].Action != "remove" {
		t.Fatalf("expected remove audit entry, got %+v", entries)
	}

	if _, err := store.Add("test", providers.ProviderConfig{Type: "demo"}, WithAuditAction("duplicate"), WithAuditDetails("from ui")); err != nil {
		t.Fatalf("duplicate: %v", err)
	}
	if entries := store.Audit(); len(entries) == 0 || entries[0].Action != "duplicate" || entries[0].Details != "from ui" {
		t.Fatalf("expected duplicate audit entry, got %+v", entries)
	}

	replacement := []ProviderRecord{
		{ID: "replacement-1", Enabled: true, Config: providers.ProviderConfig{Type: "demo", Runs: 5}},
	}
	if err := store.Replace("test", replacement); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if entries := store.Audit(); len(entries) == 0 || entries[0].Action != "replace" || entries[0].Details != "replaced 1 entries" {
		t.Fatalf("expected replace audit entry, got %+v", entries)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	reloadedEntries := reloaded.Audit()
	if len(reloadedEntries) == 0 || reloadedEntries[0].Action != "replace" {
		t.Fatalf("expected persisted replace entry, got %+v", reloadedEntries)
	}
}
