package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Migrate applies every schema migration in fsys (files named NNN_name.sql,
// sorted by name) that has not been applied yet. Each migration runs in its
// own transaction and is recorded in the schema_migrations table.
func Migrate(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	files, err := sqlFiles(fsys, ".")
	if err != nil {
		return err
	}

	if err := adoptLegacySchema(ctx, db, files); err != nil {
		return err
	}

	for _, name := range files {
		var applied int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, name).Scan(&applied); err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if applied > 0 {
			continue
		}

		script, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", name, err)
		}
	}

	return nil
}

// Seed loads every .sql file under fsys's seeds directory, in name order.
func Seed(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	files, err := sqlFiles(fsys, "seeds")
	if err != nil {
		return err
	}

	for _, name := range files {
		script, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("failed to read seed %s: %w", name, err)
		}
		if _, err := db.ExecContext(ctx, string(script)); err != nil {
			return fmt.Errorf("failed to apply seed %s: %w", name, err)
		}
	}

	return nil
}

// adoptLegacySchema marks the initial migration as applied for databases that
// were created by running 001_initial_schema.sql by hand (the old Makefile
// workflow), so it is not re-run against existing tables.
func adoptLegacySchema(ctx context.Context, db *sql.DB, files []string) error {
	if len(files) == 0 {
		return nil
	}

	var recorded, hasUsers int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&recorded); err != nil {
		return err
	}
	if recorded > 0 {
		return nil
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'users'`).Scan(&hasUsers); err != nil {
		return err
	}
	if hasUsers == 0 {
		return nil
	}

	_, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, files[0])
	return err
}

func sqlFiles(fsys fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			if dir == "." {
				files = append(files, e.Name())
			} else {
				files = append(files, dir+"/"+e.Name())
			}
		}
	}
	sort.Strings(files)
	return files, nil
}
