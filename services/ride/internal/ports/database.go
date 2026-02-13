package ports

import (
	"context"
	"database/sql"
)

type Database interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Close() error
	Ping(ctx context.Context) error
}

type Transaction interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
