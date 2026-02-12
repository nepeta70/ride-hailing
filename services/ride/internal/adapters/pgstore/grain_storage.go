package pgstore

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type GrainStorage struct {
	db *pgstore.PostgresDB
}

func NewGrainStorage(db *pgstore.PostgresDB) *GrainStorage {
	return &GrainStorage{db: db}
}

// Persist updates the grain state using the version to prevent lost updates
func (s *GrainStorage) Persist(ctx context.Context, identity *grain.GrainIdentity, data *domain.GrainData) error {
	stateJSON, err := json.Marshal(data.State)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	coreJSON, err := json.Marshal(data.Core)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	commandJSON, err := json.Marshal(data.Command)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	const query = `
		WITH inserted_event AS (
			INSERT INTO grain_events (grain_kind, grain_id, version, event_type, payload)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING grain_kind, grain_id, version
		)
		INSERT INTO grain_snapshots (grain_kind, grain_id, version, core, state)
		SELECT grain_kind, grain_id, version, $6, $7
		FROM inserted_event
		ON CONFLICT (grain_kind, grain_id) 
		DO UPDATE SET 
			version    = EXCLUDED.version,
			state      = EXCLUDED.state,
			updated_at = CURRENT_TIMESTAMP
		WHERE grain_snapshots.version < EXCLUDED.version;
    `
	result, err := s.db.ExecContext(ctx, query,
		identity.Kind,
		identity.EntityID,
		data.Version,
		data.Command.CommandName(),
		commandJSON,
		coreJSON,
		stateJSON,
	)
	if err != nil {
		return errors.NewTransientErrorf("database error during save: %w", err)
	}

	// If no rows were affected by the update (due to the WHERE clause),
	// it means a concurrency conflict occurred.
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NewTransientErrorf("concurrency violation: grain version %d is outdated", data.Version)
	}

	return nil
}

// Load retrieves the current state of the grain
func (s *GrainStorage) Load(ctx context.Context, identity *grain.GrainIdentity, target any) (int, error) {
	const query = `
		SELECT version, state 
		FROM grain_snapshots 
		WHERE grain_kind = $1 AND grain_id = $2`

	var version int
	var stateData []byte

	err := s.db.QueryRowContext(ctx, query, identity.Kind, identity.EntityID).Scan(&version, &stateData)
	if err == sql.ErrNoRows {
		return 0, errors.NewErrNotFound("grain not found")
	}
	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(stateData, target); err != nil {
		return 0, errors.NewErrJSONMarshal(err)
	}

	return version, nil
}

var _ ports.GrainStorage = (*GrainStorage)(nil)
