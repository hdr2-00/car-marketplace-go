package sqlite

import (
	"context"
	"testing"
)

// Connection-level pragmas must hold on every pooled connection, not just the
// one that happened to run a one-off PRAGMA statement.
func TestPragmasApplyToEveryPooledConnection(t *testing.T) {
	db, err := NewDatabase(*openTestDB(t))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		conn, err := db.Conn(ctx) // held open, so each iteration gets a new connection
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var fk, busy int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fk); err != nil {
			t.Fatal(err)
		}
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busy); err != nil {
			t.Fatal(err)
		}
		if fk != 1 {
			t.Errorf("connection %d: foreign_keys = %d, want 1", i, fk)
		}
		if busy <= 0 {
			t.Errorf("connection %d: busy_timeout = %d, want > 0", i, busy)
		}
	}
}
