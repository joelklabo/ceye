package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/joelklabo/ceye/internal/core"
	_ "modernc.org/sqlite"
)

// Storage provides persistent storage for historical CI run data
type Storage struct {
	db *sql.DB
}

// Config configures the storage backend
type Config struct {
	Path            string        // Path to SQLite database file
	RetentionDays   int           // How many days to keep history (0 = unlimited)
	CleanupInterval time.Duration // How often to run cleanup (0 = never)
}

// HistoricalRun wraps a core.Run with additional metadata for storage
type HistoricalRun struct {
	core.Run
	StoredAt time.Time     // When this run was stored
	WaitTime time.Duration // Time from queued to running (if applicable)
}

// New creates a new Storage instance and initializes the database
func New(cfg Config) (*Storage, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("storage path is required")
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	s := &Storage{db: db}

	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	// Start cleanup goroutine if configured
	if cfg.CleanupInterval > 0 && cfg.RetentionDays > 0 {
		go s.cleanupLoop(cfg.CleanupInterval, cfg.RetentionDays)
	}

	return s, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// migrate creates or updates the database schema
func (s *Storage) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS runs (
		id TEXT PRIMARY KEY,
		provider TEXT NOT NULL,
		repo TEXT NOT NULL,
		workflow_name TEXT NOT NULL,
		status TEXT NOT NULL,
		conclusion TEXT,
		branch TEXT,
		commit_sha TEXT,
		started_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		duration_ns INTEGER NOT NULL,
		url TEXT,
		stored_at INTEGER NOT NULL,
		wait_time_ns INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_runs_provider ON runs(provider);
	CREATE INDEX IF NOT EXISTS idx_runs_repo ON runs(repo);
	CREATE INDEX IF NOT EXISTS idx_runs_workflow ON runs(workflow_name);
	CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
	CREATE INDEX IF NOT EXISTS idx_runs_started_at ON runs(started_at);
	CREATE INDEX IF NOT EXISTS idx_runs_updated_at ON runs(updated_at);
	CREATE INDEX IF NOT EXISTS idx_runs_stored_at ON runs(stored_at);

	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	);

	INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (1, ?);
	`

	_, err := s.db.Exec(schema, time.Now().Unix())
	return err
}

// Store persists a run to the database
func (s *Storage) Store(ctx context.Context, run core.Run) error {
	hr := HistoricalRun{
		Run:      run,
		StoredAt: time.Now(),
	}
	return s.StoreHistorical(ctx, hr)
}

// StoreHistorical persists a historical run to the database
func (s *Storage) StoreHistorical(ctx context.Context, hr HistoricalRun) error {
	query := `
		INSERT OR REPLACE INTO runs (
			id, provider, repo, workflow_name, status, conclusion,
			branch, commit_sha, started_at, updated_at, duration_ns,
			url, stored_at, wait_time_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		hr.ID,
		hr.Provider,
		hr.Repo,
		hr.WorkflowName,
		hr.Status,
		hr.Conclusion,
		hr.Branch,
		hr.CommitSHA,
		hr.StartedAt.Unix(),
		hr.UpdatedAt.Unix(),
		hr.Duration.Nanoseconds(),
		hr.URL,
		hr.StoredAt.Unix(),
		hr.WaitTime.Nanoseconds(),
	)

	return err
}

// StoreBatch persists multiple runs in a single transaction
func (s *Storage) StoreBatch(ctx context.Context, runs []core.Run) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO runs (
			id, provider, repo, workflow_name, status, conclusion,
			branch, commit_sha, started_at, updated_at, duration_ns,
			url, stored_at, wait_time_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, run := range runs {
		_, err := stmt.ExecContext(ctx,
			run.ID,
			run.Provider,
			run.Repo,
			run.WorkflowName,
			run.Status,
			run.Conclusion,
			run.Branch,
			run.CommitSHA,
			run.StartedAt.Unix(),
			run.UpdatedAt.Unix(),
			run.Duration.Nanoseconds(),
			run.URL,
			now.Unix(),
			int64(0), // wait_time
		)
		if err != nil {
			return fmt.Errorf("insert run %s: %w", run.ID, err)
		}
	}

	return tx.Commit()
}

// GetRunHistory retrieves historical runs matching the criteria
func (s *Storage) GetRunHistory(ctx context.Context, provider, repo, workflow string, since time.Time, limit int) ([]HistoricalRun, error) {
	query := `
		SELECT id, provider, repo, workflow_name, status, conclusion,
		       branch, commit_sha, started_at, updated_at, duration_ns,
		       url, stored_at, wait_time_ns
		FROM runs
		WHERE 1=1
	`
	args := []interface{}{}

	if provider != "" {
		query += " AND provider = ?"
		args = append(args, provider)
	}
	if repo != "" {
		query += " AND repo = ?"
		args = append(args, repo)
	}
	if workflow != "" {
		query += " AND workflow_name = ?"
		args = append(args, workflow)
	}
	if !since.IsZero() {
		query += " AND started_at >= ?"
		args = append(args, since.Unix())
	}

	query += " ORDER BY started_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	return s.scanRuns(rows)
}

// GetProviderMetrics calculates aggregate metrics for a provider
func (s *Storage) GetProviderMetrics(ctx context.Context, provider string, since time.Time) (*Metrics, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled,
			AVG(duration_ns) as avg_duration_ns,
			MIN(duration_ns) as min_duration_ns,
			MAX(duration_ns) as max_duration_ns
		FROM runs
		WHERE provider = ? AND started_at >= ?
	`

	var m Metrics
	var avgDurNs, minDurNs, maxDurNs sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, provider, since.Unix()).Scan(
		&m.Total,
		&m.Completed,
		&m.Failed,
		&m.Cancelled,
		&avgDurNs,
		&minDurNs,
		&maxDurNs,
	)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}

	if avgDurNs.Valid {
		m.AvgDuration = time.Duration(int64(avgDurNs.Float64))
	}
	if minDurNs.Valid {
		m.MinDuration = time.Duration(int64(minDurNs.Float64))
	}
	if maxDurNs.Valid {
		m.MaxDuration = time.Duration(int64(maxDurNs.Float64))
	}

	if m.Total > 0 {
		m.SuccessRate = float64(m.Completed) / float64(m.Total)
	}

	return &m, nil
}

// GetFailureRate calculates failure rate for a repo/workflow in a time period
func (s *Storage) GetFailureRate(ctx context.Context, repo, workflow string, period time.Duration) (float64, error) {
	since := time.Now().Add(-period)

	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM runs
		WHERE repo = ? AND workflow_name = ? AND started_at >= ?
	`

	var total, failed int64
	err := s.db.QueryRowContext(ctx, query, repo, workflow, since.Unix()).Scan(&total, &failed)
	if err != nil {
		return 0, fmt.Errorf("query failure rate: %w", err)
	}

	if total == 0 {
		return 0, nil
	}

	return float64(failed) / float64(total), nil
}

// GetAverageDuration calculates average duration for a repo/workflow in a time period
func (s *Storage) GetAverageDuration(ctx context.Context, repo, workflow string, period time.Duration) (time.Duration, error) {
	since := time.Now().Add(-period)

	query := `
		SELECT AVG(duration_ns) as avg_duration_ns
		FROM runs
		WHERE repo = ? AND workflow_name = ? AND started_at >= ?
		AND status = 'completed'
	`

	var avgDurNs sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, repo, workflow, since.Unix()).Scan(&avgDurNs)
	if err != nil {
		return 0, fmt.Errorf("query average duration: %w", err)
	}

	if !avgDurNs.Valid {
		return 0, nil
	}

	return time.Duration(int64(avgDurNs.Float64)), nil
}

// Cleanup removes runs older than the specified number of days
func (s *Storage) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result, err := s.db.ExecContext(ctx, "DELETE FROM runs WHERE started_at < ?", cutoff.Unix())
	if err != nil {
		return 0, fmt.Errorf("cleanup: %w", err)
	}

	return result.RowsAffected()
}

// cleanupLoop periodically removes old runs
func (s *Storage) cleanupLoop(interval time.Duration, retentionDays int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		deleted, err := s.Cleanup(ctx, retentionDays)
		cancel()

		if err != nil {
			fmt.Printf("storage cleanup error: %v\n", err)
		} else if deleted > 0 {
			fmt.Printf("storage cleanup: deleted %d old runs\n", deleted)
		}
	}
}

// scanRuns converts database rows to HistoricalRun structs
func (s *Storage) scanRuns(rows *sql.Rows) ([]HistoricalRun, error) {
	var runs []HistoricalRun

	for rows.Next() {
		var hr HistoricalRun
		var startedAt, updatedAt, storedAt int64
		var durationNs, waitTimeNs int64

		err := rows.Scan(
			&hr.ID,
			&hr.Provider,
			&hr.Repo,
			&hr.WorkflowName,
			&hr.Status,
			&hr.Conclusion,
			&hr.Branch,
			&hr.CommitSHA,
			&startedAt,
			&updatedAt,
			&durationNs,
			&hr.URL,
			&storedAt,
			&waitTimeNs,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		hr.StartedAt = time.Unix(startedAt, 0)
		hr.UpdatedAt = time.Unix(updatedAt, 0)
		hr.StoredAt = time.Unix(storedAt, 0)
		hr.Duration = time.Duration(durationNs)
		hr.WaitTime = time.Duration(waitTimeNs)

		runs = append(runs, hr)
	}

	return runs, rows.Err()
}

// Metrics represents aggregate metrics for a set of runs
type Metrics struct {
	Total       int64
	Completed   int64
	Failed      int64
	Cancelled   int64
	SuccessRate float64
	AvgDuration time.Duration
	MinDuration time.Duration
	MaxDuration time.Duration
}
