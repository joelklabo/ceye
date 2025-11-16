package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/joelklabo/ceye/internal/providers"
)

// ProviderRecord represents a persisted entry for a runtime provider.
type ProviderRecord struct {
	ID      string                   `json:"id"`
	Enabled bool                     `json:"enabled"`
	Config  providers.ProviderConfig `json:"config"`
}

// StoreAuditEntry captures a provider store mutation for diagnostics and UI display.
type StoreAuditEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Actor        string    `json:"actor"`
	Action       string    `json:"action"`
	ProviderType string    `json:"provider_type,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
	ID           string    `json:"id,omitempty"`
	Details      string    `json:"details,omitempty"`
}

const maxAuditEntries = 200

type auditOptions struct {
	action  string
	details string
}

// AuditOption customizes how an audit entry is recorded.
type AuditOption func(*auditOptions)

// WithAuditAction overrides the default action verb stored in the audit entry.
func WithAuditAction(action string) AuditOption {
	return func(o *auditOptions) {
		o.action = action
	}
}

// WithAuditDetails adds supporting details to the audit entry.
func WithAuditDetails(details string) AuditOption {
	return func(o *auditOptions) {
		o.details = details
	}
}

func applyAuditOptions(defaultAction, defaultDetails string, opts []AuditOption) (string, string) {
	res := auditOptions{action: defaultAction, details: defaultDetails}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&res)
	}
	if res.action == "" {
		res.action = defaultAction
	}
	return res.action, res.details
}

type persistentStore struct {
	Entries []ProviderRecord  `json:"entries"`
	Audit   []StoreAuditEntry `json:"audit,omitempty"`
}

// Store manages provider records stored on disk.
type Store struct {
	path    string
	entries []ProviderRecord
	audit   []StoreAuditEntry
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

// Audit returns the most recent provider store mutation entries (newest first).
func (s *Store) Audit() []StoreAuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]StoreAuditEntry, len(s.audit))
	copy(records, s.audit)
	return records
}

// Replace replaces the store entries with the provided list, writing new data.
func (s *Store) Replace(actor string, entries []ProviderRecord, opts ...AuditOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prevEntries := append([]ProviderRecord(nil), s.entries...)
	s.entries = append([]ProviderRecord(nil), entries...)
	defaultDetails := fmt.Sprintf("replaced %d entries", len(entries))
	action, details := applyAuditOptions("replace", defaultDetails, opts)
	record := ProviderRecord{Config: providers.ProviderConfig{DisplayName: "provider store"}}
	revert := s.recordAuditLocked(actor, action, record, details)
	if err := s.save(); err != nil {
		s.entries = prevEntries
		revert()
		return err
	}
	return nil
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
func (s *Store) Add(actor string, cfg providers.ProviderConfig, opts ...AuditOption) (ProviderRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := ProviderRecord{ID: uuid.NewString(), Enabled: true, Config: cfg}
	s.entries = append(s.entries, record)
	action, details := applyAuditOptions("add", "", opts)
	revert := s.recordAuditLocked(actor, action, record, details)
	if err := s.save(); err != nil {
		revert()
		return ProviderRecord{}, err
	}
	return record, nil
}

// Update replaces the configuration for an existing record.
func (s *Store) Update(actor, id string, cfg providers.ProviderConfig, opts ...AuditOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	s.entries[idx].Config = cfg
	action, details := applyAuditOptions("update", "", opts)
	entry := s.entries[idx]
	revert := s.recordAuditLocked(actor, action, entry, details)
	if err := s.save(); err != nil {
		revert()
		return err
	}
	return nil
}

// SetEnabled toggles the enabled state of a record.
func (s *Store) SetEnabled(actor, id string, enabled bool, opts ...AuditOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	s.entries[idx].Enabled = enabled
	defaultAction := "disable"
	if enabled {
		defaultAction = "enable"
	}
	action, details := applyAuditOptions(defaultAction, "", opts)
	entry := s.entries[idx]
	revert := s.recordAuditLocked(actor, action, entry, details)
	if err := s.save(); err != nil {
		revert()
		return err
	}
	return nil
}

// Remove deletes a record from the store.
func (s *Store) Remove(actor, id string, opts ...AuditOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index(id)
	if idx < 0 {
		return fmt.Errorf("provider %s not found", id)
	}
	entry := s.entries[idx]
	s.entries = append(s.entries[:idx], s.entries[idx+1:]...)
	action, details := applyAuditOptions("remove", "", opts)
	revert := s.recordAuditLocked(actor, action, entry, details)
	if err := s.save(); err != nil {
		// restore removed entry on failure
		s.entries = append(s.entries[:idx], append([]ProviderRecord{entry}, s.entries[idx:]...)...)
		revert()
		return err
	}
	return nil
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
	s.audit = append([]StoreAuditEntry(nil), raw.Audit...)
	return nil
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	out := persistentStore{Entries: s.entries, Audit: s.audit}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("encode provider store: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("persist provider store: %w", err)
	}
	return nil
}

func (s *Store) recordAuditLocked(actor, action string, record ProviderRecord, details string) func() {
	if actor == "" {
		actor = "system"
	}
	entry := StoreAuditEntry{
		Timestamp:    time.Now(),
		Actor:        actor,
		Action:       action,
		ProviderType: record.Config.Type,
		DisplayName:  providers.DisplayName(record.Config),
		ID:           record.ID,
		Details:      details,
	}
	s.audit = append([]StoreAuditEntry{entry}, s.audit...)
	if len(s.audit) > maxAuditEntries {
		s.audit = s.audit[:maxAuditEntries]
	}
	return func() {
		if len(s.audit) == 0 {
			return
		}
		if s.audit[0] == entry {
			s.audit = s.audit[1:]
		}
	}
}
