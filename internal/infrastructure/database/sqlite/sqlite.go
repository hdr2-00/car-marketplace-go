package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Config holds SQLite database configuration
type Config struct {
	DatabasePath    string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabase creates a new SQLite database connection
func NewDatabase(config Config) (*sql.DB, error) {
	// Open database connection. Pragmas go in the DSN so the driver applies
	// them to every pooled connection, not just the first one.
	db, err := sql.Open("sqlite", dsn(config.DatabasePath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// dsn builds the connection string for modernc.org/sqlite:
//   - foreign_keys: enforce FKs and ON DELETE actions (off by default in SQLite)
//   - busy_timeout: wait for the write lock instead of failing with SQLITE_BUSY
//   - journal_mode=WAL: readers don't block the writer
//   - _txlock=immediate: take the write lock at BEGIN, avoiding lock-upgrade deadlocks
//   - _time_format=sqlite: store times in a sortable format SQLite date functions understand
func dsn(path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + strings.Join([]string{
		"_pragma=foreign_keys(1)",
		"_pragma=busy_timeout(5000)",
		"_pragma=journal_mode(WAL)",
		"_txlock=immediate",
		"_time_format=sqlite",
	}, "&")
}

// DefaultConfig returns default SQLite configuration
func DefaultConfig(dbPath string) Config {
	return Config{
		DatabasePath:    dbPath,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}
