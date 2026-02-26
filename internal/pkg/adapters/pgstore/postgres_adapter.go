package pgstore

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

const postgresServiceName = "PostgresDB"

type PostgresOpts struct {
	Config         *PostgresConfig
	RetrierFactory ports.RetrierFactoryInterface
	Telemetry      ports.TelemetryProvider
}

func (o *PostgresOpts) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("PostgresConfig is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("Telemetry is required")
	}
	if o.RetrierFactory == nil {
		return errors.NewValidationErrorf("RetrierFactory is required")
	}
	return nil
}

type PostgresDB struct {
	config       *PostgresConfig
	retryFactory ports.RetrierFactoryInterface
	telemetry    ports.TelemetryProvider
	*sql.DB
}

func NewPostgresDB(opts *PostgresOpts) (*PostgresDB, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	dbConn, err := sql.Open("postgres", opts.Config.DSN())
	if err != nil {
		// If the driver name or DSN format is wrong, it's a permanent code/config bug
		opts.Telemetry.Logger().Error("Failed to open postgres connection", "error", err)
		return nil, errors.NewPermanentErrorf("failed to open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Config.PingTimeout)
	defer cancel()

	retrier := opts.RetrierFactory.NewExponentialBackoffRetrier(postgresServiceName, opts.Config.PingTimeout)
	err = retrier.Do(ctx, func() error {
		return dbConn.PingContext(ctx)
	})

	if err != nil {
		opts.Telemetry.Logger().Error("Failed to ping postgres after retries", "error", err)
		return nil, errors.NewPermanentErrorf("failed to ping postgres: %w", err)
	}

	db := &PostgresDB{
		config:       opts.Config,
		telemetry:    opts.Telemetry,
		retryFactory: opts.RetrierFactory,
		DB:           dbConn,
	}
	return db, nil
}

func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Transaction, error) {
	beginCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()
	tx, err := db.DB.BeginTx(beginCtx, opts)
	if err != nil {
		db.telemetry.Logger().Error("Failed to begin postgres transaction", "error", err)
		return nil, errors.NewTransientErrorf("failed to begin postgres transaction: %w", err)
	}
	return &PostgresTx{tx}, nil
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.PingTimeout)
	defer cancel()
	if err := db.DB.PingContext(ctx); err != nil {
		// Database being unreachable is a transient infrastructure issue
		db.telemetry.Logger().DebugContext(ctx, "Postgres health check failed", "error", err)
		db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "health_check", "error")
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
	tracer := db.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, "DB Query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.statement", query),
		),
	)
	defer span.End()

	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	db.telemetry.Logger().Debug("Executing query with retry", "query", query)
	var rows *sql.Rows
	err := strategy.Do(ctx, func() error {
		var err error
		rows, err = db.DB.QueryContext(ctx, query, args...)
		if err != nil {
			db.telemetry.Logger().ErrorContext(ctx, "Query execution failed", "error", err)
			// You can add logic here to check if the error is worth retrying
			db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "query", "error")
			return errors.NewTransientErrorf("query failed: %w", err)
		}
		return nil
	})
	db.telemetry.Logger().Debug("Query returned rows", "rows", rows, "error", err)
	return rows, err
}

// ExecContext follows the same pattern for INSERT/UPDATE/DELETE
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tracer := db.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, "DB Exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.String("db.statement", query),
		),
	)
	defer span.End()
	var res sql.Result

	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	err := strategy.Do(ctx, func() error {
		var err error
		res, err = db.DB.ExecContext(ctx, query, args...)
		if err != nil {
			db.telemetry.Logger().ErrorContext(ctx, "Exec execution failed", "error", err)
			db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "exec", "error")
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
