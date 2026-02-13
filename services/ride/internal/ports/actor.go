package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
)

type GrainStorage interface {
	Persist(ctx context.Context, identity *grain.GrainIdentity, data *domain.GrainData) error
	Load(ctx context.Context, identity *grain.GrainIdentity, target any) (int, error)
}

type GrainSystemInterface interface {
	Silo() pkgPorts.Silo
	GrainIdentityFactory() *service.GrainIdentityFactory
	GrainPersistence() GrainStorage
}
