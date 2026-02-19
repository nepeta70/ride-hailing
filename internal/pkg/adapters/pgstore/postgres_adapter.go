package pgstore

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

const postgresServiceName = "PostgresDB"

type PostgresDB struct {
	config       *PostgresConfig
	logger       ports.Logger
	retryFactory ports.RetrierFactoryInterface
	*sql.DB
}

func NewPostgresDB(config *PostgresConfig, retrierFactory ports.RetrierFactoryInterface, logger ports.Logger) (*PostgresDB, error) {
	dbConn, err := sql.Open("postgres", config.DSN())
	if err != nil {
		// If the driver name or DSN format is wrong, it's a permanent code/config bug
		return nil, errors.NewPermanentErrorf("failed to open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.PingTimeout)
	defer cancel()

	strategy := retrierFactory.NewExponentialBackoffRetrier(postgresServiceName, config.PingTimeout)
	err = strategy.Do(ctx, func() error {
		return dbConn.PingContext(ctx)
	})

	if err != nil {
		return nil, errors.NewPermanentErrorf("failed to ping postgres: %w", err)
	}

	db := &PostgresDB{
		config:       config,
		logger:       logger,
		retryFactory: retrierFactory,
		DB:           dbConn,
	}
	return db, nil
}

func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Transaction, error) {
	beginCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()
	tx, err := db.DB.BeginTx(beginCtx, opts)
	if err != nil {
		return nil, errors.NewTransientErrorf("failed to begin postgres transaction: %w", err)
	}
	return &PostgresTx{tx}, nil
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.PingTimeout)
	defer cancel()
	if err := db.DB.PingContext(ctx); err != nil {
		// Database being unreachable is a transient infrastructure issue
		db.logger.Error("Postgres health check failed", "error", err)
		return errors.NewTransientErrorf("postgres ping failed: %w", err)
	}
	return nil
}

func (db *PostgresDB) ServiceName() string {
	return postgresServiceName
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// QueryContext wraps the standard sql.DB QueryContext with retries and timeouts
func (db *PostgresDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	db.logger.Debug("Executing query with retry", "query", query)
	var rows *sql.Rows
	err := strategy.Do(ctx, func() error {
		var err error
		rows, err = db.DB.QueryContext(ctx, query, args...)
		if err != nil {
			// You can add logic here to check if the error is worth retrying
			return errors.NewTransientErrorf("query failed: %w", err)
		}
		return nil
	})
	db.logger.Debug("Query returned rows", "rows", rows, "error", err)
	return rows, err
}

// ExecContext follows the same pattern for INSERT/UPDATE/DELETE
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var res sql.Result

	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	err := strategy.Do(ctx, func() error {
		var err error
		res, err = db.DB.ExecContext(ctx, query, args...)
		if err != nil {
			return errors.NewTransientErrorf("exec failed: %w", err)
		}
		return nil
	})

	return res, err
}

// PostgresTx wraps *sql.Tx to satisfy ports.Transaction
type PostgresTx struct {
	*sql.Tx
}

// Ensure interface compliance at compile time
var _ ports.Database = (*PostgresDB)(nil)
var _ ports.Transaction = (*PostgresTx)(nil)
