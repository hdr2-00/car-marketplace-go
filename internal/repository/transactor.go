package repository

import "context"

// Transactor runs a function inside a database transaction. Repository calls
// made with the context passed to fn take part in that transaction; the
// transaction commits if fn returns nil and rolls back otherwise.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
