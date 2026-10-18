package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"car-marketplace/internal/repository"
)

// querier is the subset of *sql.DB and *sql.Tx that repositories use.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type txKey struct{}

// conn returns the transaction carried by ctx, or db when there is none.
func conn(ctx context.Context, db *sql.DB) querier {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

// withTx runs fn in the transaction carried by ctx, or in a new one.
func withTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, q querier) error) error {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx, tx)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(context.WithValue(ctx, txKey{}, tx), tx); err != nil {
		return err
	}
	return tx.Commit()
}

type transactor struct {
	db *sql.DB
}

// NewTransactor creates a repository.Transactor backed by db.
func NewTransactor(db *sql.DB) repository.Transactor {
	return &transactor{db: db}
}

// WithinTx runs fn in a transaction; nested calls join the outer transaction.
func (t *transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return withTx(ctx, t.db, func(ctx context.Context, _ querier) error {
		return fn(ctx)
	})
}

func isUniqueViolation(err error, column string) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") &&
		(column == "" || strings.Contains(err.Error(), column))
}

func isForeignKeyViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}

// nullString stores empty strings as NULL, so optional UNIQUE columns
// (such as a listing's VIN) don't collide on the empty string.
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
