package app

import (
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/silo"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type GrainSystem struct {
	si                   *silo.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	grainPersistence     ports.GrainStorage
	eventPub             pkgPorts.EventPublisher
	logger               pkgPorts.Logger
	topic                contracts.Topic
	processMu            sync.Mutex // prevents concurrent ProcessWaitlist execution
}

func NewGrainSystem(config *config.BaseConfig, storage ports.GrainStorage, eventStore pkgPorts.EventStore, eventPub pkgPorts.EventPublisher, logger pkgPorts.Logger) *GrainSystem {
	topic := contracts.TopicRide

	si := silo.NewSilo(&config.Timeouts, logger)
	si.RegisterFactory(domain.RideGrainKind, func(identity *grain.GrainIdentity) pkgPorts.Grain {
		return grains.NewRideGrain(storage, eventPub, logger, topic)
	})

	return &GrainSystem{
		si:                   si,
		eventPub:             eventPub,
		logger:               logger,
		topic:                topic,
		grainPersistence:     storage,
		grainIdentityFactory: &service.GrainIdentityFactory{},
	}
}

func (gs *GrainSystem) Silo() pkgPorts.Silo {
	return gs.si
}

func (gs *GrainSystem) GrainIdentityFactory() *service.GrainIdentityFactory {
	return gs.grainIdentityFactory
}

func (gs *GrainSystem) GrainPersistence() ports.GrainStorage {
	return gs.grainPersistence
}

var _ ports.GrainSystemInterface = (*GrainSystem)(nil)
