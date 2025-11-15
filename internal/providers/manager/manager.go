package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"github.com/joelklabo/ceye/internal/providers"
)

// ProviderRecord represents a persisted entry for a runtime provider.
type ProviderRecord struct {
	ID      string                   `json:"id"`
	Enabled bool                     `json:"enabled"`
	Config  providers.ProviderConfig `json:"config"`
}

type persistentStore struct {
	Entries []ProviderRecord `json:"entries"`
}

// Store manages provider records stored on disk.
type Store struct {
	path    string
	entries []ProviderRecord
	mu      sync.RWMutex
}

// New creates a Store that persists to path. Path must not be empty.
func New(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("provider store path is required")
	}
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// List returns a copy of all provider records regardless of enabled state.
func (s *Store) List() []ProviderRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]ProviderRecord, len(s.entries))
	copy(records, s.entries)
	return records
}

// EnabledRecords returns the records that are currently enabled.
func (s *Store) EnabledRecords() []ProviderRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	enabled := make([]ProviderRecord, 0, len(s.entries))
	for _, r := range s.entries {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled
}

// Find locates a record by ID.
func (s *Store) Find(id string) (ProviderRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.entries {
		if r.ID == id {
			return r, true
		}
	}
	return ProviderRecord{}, false
}

// Add adds a new provider record and persists the store.
func (s *Store) Add(cfg providers.ProviderConfig) (ProviderRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := ProviderRecord{ID: uuid.NewString(), Enabled: true, Config: cfg}
	s.entries = append(s.entries, record)
	if err := s.save(); err != nil {
		return ProviderRecord{}, err
	}
	return record, nil
}

// Update replaces the configuration for an existing record.
func (s *Store) Update(id string, cfg providers.ProviderConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	s.entries[idx].Config = cfg
	return s.save()
}

// SetEnabled toggles the enabled state of a record.
func (s *Store) SetEnabled(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	s.entries[idx].Enabled = enabled
	return s.save()
}

// Remove deletes a record from the store.
func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	s.entries = append(s.entries[:idx], s.entries[idx+1:]...)
	return s.save()
}

func (s *Store) index(id string) int {
	for i, r := range s.entries {
		if r.ID == id {
			return i
		}
	}
	return -1
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.entries = nil
		return nil
	}
	if err != nil {
		return fmt.Errorf("read provider store: %w", err)
	}
	var raw persistentStore
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode provider store: %w", err)
	}
	s.entries = raw.Entries
	return nil
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	out := persistentStore{Entries: s.entries}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("encode provider store: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("persist provider store: %w", err)
	}
	return nil
}
