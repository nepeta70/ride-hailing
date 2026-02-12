package actor

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

// GrainStorage provides shared event persistence for grains
type GrainStorage struct {
	db         ports.Database
	eventStore ports.EventStore
}

func NewGrainStorage(db ports.Database, eventStore ports.EventStore) *GrainStorage {
	return &GrainStorage{db: db, eventStore: eventStore}
}

// PersistEvent handles transaction creation, event append, and commit
func (s *GrainStorage) PersistEvent(ctx context.Context, event *contracts.EventMessage) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := s.eventStore.Append(ctx, tx, event); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetStream loads the event stream for an aggregate
func (s *GrainStorage) GetStream(ctx context.Context, streamId enums.AggregateType, ID string) ([]contracts.EventMessage, error) {
	return s.eventStore.GetStream(ctx, streamId, ID)
}
