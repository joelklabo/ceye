package core

// Store aggregates Run updates from providers.
type Store struct{}

// NewStore constructs an empty Store.
func NewStore() *Store {
	return &Store{}
}

// Merge ingests a RunEvent into the Store state.
func (s *Store) Merge(event RunEvent) {
	// TODO: implement merging logic.
}

// ListRuns returns runs optionally filtered by provider.
func (s *Store) ListRuns(providerFilter string) []Run {
	return nil
}
