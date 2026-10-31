// Package migrations embeds the SQL schema migrations and seed files so the
// application can apply them itself, without the sqlite3 CLI.
package migrations

import "embed"

// Schema holds the numbered schema migrations (NNN_name.sql), applied in order.
//
//go:embed *.sql
var Schema embed.FS

// Seeds holds optional sample data, loaded only on request.
//
//go:embed seeds/*.sql
var Seeds embed.FS
