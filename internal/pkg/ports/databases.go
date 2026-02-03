package ports

import (
	"context"
	"database/sql"
)

type database interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Close() error
}

type Database interface {
	database
	HealthProvider
}

type Transaction interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
