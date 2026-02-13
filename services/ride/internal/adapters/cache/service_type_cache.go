package cache

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type ServiceTypeCache struct {
	readRepo ports.ServiceTypeReadRepository
	store    atomic.Pointer[map[string]*domain.ServiceType]
	once     sync.Once
}

func NewServiceTypeCache(serviceTypeReadRepo ports.ServiceTypeReadRepository) *ServiceTypeCache {
	return &ServiceTypeCache{
		readRepo: serviceTypeReadRepo,
	}
}

func (c *ServiceTypeCache) GetServiceTypeByCode(ctx context.Context, code string) (*domain.ServiceType, bool) {
	c.once.Do(func() {
		_ = c.Refresh(ctx)
	})

	data := c.store.Load()
	if data == nil {
		return nil, false
	}

	serviceType, ok := (*data)[code]
	return serviceType, ok
}

func (c *ServiceTypeCache) Refresh(ctx context.Context) error {
	// store, err := c.readRepo.GetAllEnabled(ctx) // TDOO: Implement this method to fetch data from the repository
	// if err != nil {
	// 	return err
	// }
	store := make(map[string]*domain.ServiceType, 1)
	store["STANDARD"] = &domain.ServiceType{
		Code:          "STANDARD",
		Name:          "Standard ride",
		MaxPassengers: 3,
	}

	c.store.Store(&store)
	return nil
}

var _ ports.ServiceTypeCacheInterface = (*ServiceTypeCache)(nil)
