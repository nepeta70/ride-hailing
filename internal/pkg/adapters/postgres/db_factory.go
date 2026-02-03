package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	retry "github.com/nepeta70/ride-hailing/internal/pkg/resiliency"
)

// PostgresDbFactory handles initialization of Postgres database and related components
type PostgresDbFactory struct {
	config *PostgresConfig
	db     *sql.DB
}

// NewPostgresDbFactory creates a new PostgresDbFactory with injected config
func NewPostgresDbFactory(config *PostgresConfig) (*PostgresDbFactory, error) {
	factory := &PostgresDbFactory{
		config: config,
	}

	// Initialize the database connection
	rawDB, err := factory.initDB()
	if err != nil {
		return nil, err
	}

	factory.db = rawDB
	return factory, nil
}

// CreateDatabase returns a Database instance
func (f *PostgresDbFactory) CreateDatabase() ports.Database {
	return &PostgresDB{DB: f.db}
}

// initDB establishes a connection to Postgres and validates it
func (f *PostgresDbFactory) initDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", f.config.DSN())
	if err != nil {
		// If the driver name or DSN format is wrong, it's a permanent code/config bug
		return nil, errors.NewPermanentErrorf("failed to open postgres connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Use your custom strategy
	strategy := retry.NewExponentialBackoff(retry.DefaultConfig())

	// Use your generic Do function
	err = retry.Do(ctx, strategy, func() error {
		return db.Ping()
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}
