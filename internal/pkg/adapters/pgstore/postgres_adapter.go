package pgstore

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

const postgresServiceName = "PostgresDB"

type PostgresOpts struct {
	Config         *PostgresConfig
	RetrierFactory ports.RetrierFactoryInterface
	Telemetry      ports.TelemetryProvider
	ContextManager *ctxmgr.ContextManager
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
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("ContextManager is required")
	}
	return nil
}

type PostgresDB struct {
	config       *PostgresConfig
	retryFactory ports.RetrierFactoryInterface
	telemetry    ports.TelemetryProvider
	ctxManager   *ctxmgr.ContextManager
	*sql.DB
}

func NewPostgresDB(opts *PostgresOpts) (*PostgresDB, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	ctx, span := opts.Telemetry.Tracer().Start(context.Background(), "Postgres:Initialize",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.Bool("service.init", true)),
	)
	defer span.End()

	dbConn, err := sql.Open("postgres", opts.Config.DSN())
	if err != nil {
		// If the driver name or DSN format is wrong, it's a permanent code/config bug
		opts.Telemetry.Logger().ErrorContext(ctx, "Failed to open postgres connection", "error", err)
		span.SetStatus(codes.Error, "failed to open connection")
		return nil, errors.NewPermanentErrorf("failed to open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Config.PingTimeout)
	defer cancel()

	retrier := opts.RetrierFactory.NewExponentialBackoffRetrier(postgresServiceName, opts.Config.PingTimeout)
	err = retrier.Do(ctx, func(ctx context.Context) error {
		return dbConn.PingContext(ctx)
	})

	if err != nil {
		opts.Telemetry.Logger().ErrorContext(ctx, "Failed to ping postgres after retries", "error", err)
		span.SetStatus(codes.Error, "failed to ping after retries")
		return nil, errors.NewPermanentErrorf("failed to ping postgres: %w", err)
	}

	db := &PostgresDB{
		config:       opts.Config,
		telemetry:    opts.Telemetry,
		retryFactory: opts.RetrierFactory,
		ctxManager:   opts.ContextManager,
		DB:           dbConn,
	}
	return db, nil
}

func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Transaction, error) {
	span := db.TraceSpan(ctx, "Begin Postgres Transaction")
	defer span.End()

	beginCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()
	tx, err := db.DB.BeginTx(beginCtx, opts)
	if err != nil {
		db.telemetry.Logger().ErrorContext(ctx, "Failed to begin postgres transaction", "error", err)
		db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "begin_tx", "error")
		span.SetStatus(codes.Error, "failed to begin transaction")
		return nil, errors.NewTransientErrorf("failed to begin postgres transaction: %w", err)
	}
	return &PostgresTx{tx}, nil
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.PingTimeout)
	defer cancel()
	if err := db.DB.PingContext(ctx); err != nil {
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
	span := db.TraceSpan(ctx, "DB Query")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
	)

	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	db.telemetry.Logger().DebugContext(ctx, "Executing query with retry", "query", query)
	var rows *sql.Rows
	err := strategy.Do(ctx, func(ctx context.Context) error {
		var err error
		rows, err = db.DB.QueryContext(ctx, query, args...)
		if err != nil {
			db.telemetry.Logger().ErrorContext(ctx, "Query execution failed", "error", err)
			// You can add logic here to check if the error is worth retrying
			db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "query", "error")
			span.SetStatus(codes.Error, "query execution failed")
			return errors.NewTransientErrorf("query failed: %w", err)
		}
		return nil
	})
	db.telemetry.Logger().DebugContext(ctx, "Query returned rows", "rows", rows, "error", err)
	return rows, err
}

// ExecContext follows the same pattern for INSERT/UPDATE/DELETE
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	span := db.TraceSpan(ctx, "DB Exec")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
	)
	var res sql.Result

	strategy := db.retryFactory.NewExponentialBackoffRetrier(postgresServiceName, db.config.QueryTimeout)

	err := strategy.Do(ctx, func(ctx context.Context) error {
		var err error
		res, err = db.DB.ExecContext(ctx, query, args...)
		if err != nil {
			db.telemetry.Logger().ErrorContext(ctx, "Exec execution failed", "error", err)
			db.telemetry.Metrics().DependencyFailure(db.ServiceName(), "exec", "error")
			span.SetStatus(codes.Error, "exec execution failed")
			return errors.NewTransientErrorf("exec failed: %w", err)
		}
		return nil
	})

	return res, err
}

func (db *PostgresDB) TraceSpan(ctx context.Context, spanName string) trace.Span {
	tracer := db.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgres"),
		),
	)

	info, ok := db.ctxManager.Extract(ctx)
	if ok {
		span.SetAttributes(
			attribute.String("sender.id", info.Sender.ID.String()),
			attribute.String("request.id", info.Trace.RequestID.String()),
		)
	}
	return span
}

// PostgresTx wraps *sql.Tx to satisfy ports.Transaction
type PostgresTx struct {
	*sql.Tx
}

// Ensure interface compliance at compile time
var _ ports.Database = (*PostgresDB)(nil)
var _ ports.Transaction = (*PostgresTx)(nil)
