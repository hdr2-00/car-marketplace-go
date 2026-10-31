package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"car-marketplace/migrations"
)

func openTestDB(t *testing.T) *Config {
	t.Helper()
	cfg := DefaultConfig(filepath.Join(t.TempDir(), "test.db"))
	return &cfg
}

func TestMigrateAppliesSchemaOnFreshDatabase(t *testing.T) {
	cfg := openTestDB(t)
	db, err := NewDatabase(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := Migrate(context.Background(), db, migrations.Schema); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_types`).Scan(&n); err != nil {
		t.Fatalf("schema not applied: %v", err)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	cfg := openTestDB(t)
	db, err := NewDatabase(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := Migrate(context.Background(), db, migrations.Schema); err != nil {
			t.Fatalf("migrate run %d: %v", i+1, err)
		}
	}
}

func TestMigrateAdoptsDatabaseCreatedByMakefile(t *testing.T) {
	cfg := openTestDB(t)
	db, err := NewDatabase(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Simulate the old `sqlite3 db < 001_initial_schema.sql` workflow:
	// schema present, but no migration bookkeeping table.
	initial, err := migrations.Schema.ReadFile("001_initial_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(initial)); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(context.Background(), db, migrations.Schema); err != nil {
		t.Fatalf("migrate legacy db: %v", err)
	}
}

func TestSeedLoadsSampleData(t *testing.T) {
	cfg := openTestDB(t)
	db, err := NewDatabase(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := Migrate(ctx, db, migrations.Schema); err != nil {
		t.Fatal(err)
	}
	if err := Seed(ctx, db, migrations.Seeds); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM car_listings`).Scan(&n); err != nil || n == 0 {
		t.Fatalf("expected seeded listings, got %d (%v)", n, err)
	}
}
