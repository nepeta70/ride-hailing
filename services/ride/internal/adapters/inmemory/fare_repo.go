package inmemory

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

const bufferFares = 1000

type InMemoryFareRepo struct {
	data map[uuid.UUID]*domain.Fares
}

func NewInMemoryFareRepo() *InMemoryFareRepo {
	return &InMemoryFareRepo{
		data: make(map[uuid.UUID]*domain.Fares, bufferFares),
	}
}

func (repo *InMemoryFareRepo) Save(ctx context.Context, fare *domain.Fares) error {
	if _, exists := repo.data[fare.ID]; exists {
		return nil
	}
	repo.data[fare.ID] = fare
	return nil
}

func (repo *InMemoryFareRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fares, error) {
	fare, exists := repo.data[id]
	if !exists {
		return nil, nil // or return a domain.ErrFareNotFound if defined
	}
	return fare, nil
}

var _ ports.FareReadRepository = (*InMemoryFareRepo)(nil)
var _ ports.FareWriteRepository = (*InMemoryFareRepo)(nil)
