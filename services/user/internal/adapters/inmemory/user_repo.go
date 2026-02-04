package inmemory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[uuid.UUID]*domain.User, 1000),
	}
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID()] = user
	return nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID()]; !ok {
		return domain.ErrUserNotFound
	}
	r.users[user.ID()] = user
	return nil
}

var _ ports.ReadUserRepository = (*InMemoryUserRepository)(nil)
var _ ports.WriteUserRepository = (*InMemoryUserRepository)(nil)
