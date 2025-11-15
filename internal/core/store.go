package core

import (
	"fmt"
	"sort"
	"sync"
)

// Store aggregates Run updates from providers in a thread-safe map.
type Store struct {
	mu   sync.RWMutex
	runs map[string]Run
}

// NewStore constructs an empty Store.
func NewStore() *Store {
	return &Store{
		runs: make(map[string]Run),
	}
}

// Merge ingests a RunEvent into the Store state.
func (s *Store) Merge(event RunEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.Err != nil {
		return
	}

	for _, run := range event.Runs {
		if run.Provider == "" {
			run.Provider = event.Provider
		}
		if run.ID == "" {
			// Fall back to URL for uniqueness if no ID present.
			run.ID = run.URL
		}
		if run.ID == "" {
			// If still empty, skip to avoid corrupting the map.
			continue
		}
		key := storeKey(run.Provider, run.ID)
		s.runs[key] = run
	}
}

// ListRuns returns runs optionally filtered by provider. Runs are sorted by
// UpdatedAt descending (newest first) and then by ID for determinism.
func (s *Store) ListRuns(providerFilter string) []Run {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter := providerFilter
	out := make([]Run, 0, len(s.runs))
	for _, run := range s.runs {
		if filter != "" && run.Provider != filter {
			continue
		}
		out = append(out, run)
	}

	sort.Slice(out, func(i, j int) bool {
		switch {
		case out[i].UpdatedAt.After(out[j].UpdatedAt):
			return true
		case out[i].UpdatedAt.Before(out[j].UpdatedAt):
			return false
		default:
			return out[i].ID > out[j].ID
		}
	})

	return out
}

func storeKey(provider, id string) string {
	return fmt.Sprintf("%s:%s", provider, id)
}
