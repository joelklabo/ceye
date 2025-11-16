package core

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// StorageBackend defines the interface for persistent storage
type StorageBackend interface {
	Store(ctx context.Context, run Run) error
	StoreBatch(ctx context.Context, runs []Run) error
}

// Store aggregates Run updates from providers in a thread-safe map.
type Store struct {
	mu      sync.RWMutex
	runs    map[string]Run
	storage StorageBackend // Optional persistent storage
}

// NewStore constructs an empty Store.
func NewStore() *Store {
	return &Store{
		runs:    make(map[string]Run),
		storage: nil,
	}
}

// NewStoreWithStorage creates a store with persistent storage backend
func NewStoreWithStorage(storage StorageBackend) *Store {
	return &Store{
		runs:    make(map[string]Run),
		storage: storage,
	}
}

// Merge ingests a RunEvent into the Store state.
func (s *Store) Merge(event RunEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.Err != nil {
		return
	}

	log.Printf("STORE: merge called - %d runs from provider '%s'", len(event.Runs), event.Provider)

	var completedRuns []Run

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
		
		// Track completed runs for storage
		if s.storage != nil && run.Status == RunStatusCompleted {
			oldRun, existed := s.runs[key]
			// Only persist if newly completed (not already completed)
			if !existed || oldRun.Status != RunStatusCompleted {
				completedRuns = append(completedRuns, run)
			}
		}
		
		s.runs[key] = run
	}

	log.Printf("STORE: now contains %d total runs", len(s.runs))

	// Persist completed runs asynchronously (don't block on storage)
	if len(completedRuns) > 0 && s.storage != nil {
		go func(runs []Run) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := s.storage.StoreBatch(ctx, runs); err != nil {
				// Log error but don't fail the merge
				// In a production system, you'd want to emit a metric here
				_ = err
			}
		}(completedRuns)
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
		si, sj := statusRank(out[i]), statusRank(out[j])
		if si != sj {
			return si < sj
		}
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

func statusRank(run Run) int {
	switch run.Status {
	case RunStatusInProgress:
		return 0
	case RunStatusQueued:
		return 1
	case RunStatusFailed, RunStatusCancelled:
		return 2
	case RunStatusCompleted:
		if strings.EqualFold(run.Conclusion, "success") || strings.EqualFold(run.Conclusion, "succeeded") || run.Conclusion == "" {
			return 3
		}
		return 2
	default:
		return 4
	}
}
