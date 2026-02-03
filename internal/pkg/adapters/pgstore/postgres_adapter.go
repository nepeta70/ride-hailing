package pgstore

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	retry "github.com/nepeta70/ride-hailing/internal/pkg/resiliency"
)

type PostgresDB struct {
	config *PostgresConfig
	*sql.DB
}

func NewPostgresDB(config *PostgresConfig) (*PostgresDB, error) {
	dbConn, err := sql.Open("postgres", config.DSN())
	if err != nil {
		// If the driver name or DSN format is wrong, it's a permanent code/config bug
		return nil, errors.NewPermanentErrorf("failed to open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.PingTimeout)
	defer cancel()

	strategy := retry.NewExponentialBackoff(retry.DefaultConfig())

	err = retry.Do(ctx, strategy, func() error {
		return dbConn.PingContext(ctx)
	})

	if err != nil {
		return nil, errors.NewTransientErrorf("failed to ping postgres: %w", err)
	}

	db := &PostgresDB{
		config: config,
		DB:     dbConn,
	}
	return db, nil
}

func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Transaction, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, errors.NewTransientErrorf("failed to begin postgres transaction: %w", err)
	}
	return &PostgresTx{tx}, nil
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	if err := db.DB.PingContext(ctx); err != nil {
		// Database being unreachable is a transient infrastructure issue
		return errors.NewTransientErrorf("postgres ping failed: %w", err)
	}
	return nil
}

func (db *PostgresDB) ServiceName() string {
	return "Postgres"
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// PostgresTx wraps *sql.Tx to satisfy ports.Transaction
type PostgresTx struct {
	*sql.Tx
}

// Ensure interface compliance at compile time
var _ ports.Database = (*PostgresDB)(nil)
var _ ports.Transaction = (*PostgresTx)(nil)
