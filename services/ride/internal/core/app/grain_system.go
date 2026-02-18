package app

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/silo"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type GrainSystemOptions struct {
	Storage        ports.GrainStorage
	EventPublisher pkgPorts.EventPublisher
	Logger         pkgPorts.Logger
	Topic          contracts.Topic
	GrainTimeout   time.Duration
	RetrierFactory *retry.RetrierFactory
}

func (opts *GrainSystemOptions) Validate() error {
	if opts.Storage == nil {
		return errors.NewValidationErrorf("grain storage is required")
	}
	// TODO: uncomment when event publisher is implemented
	// if opts.EventPublisher == nil {
	// 	return errors.NewValidationErrorf("event publisher is required")
	// }
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	if opts.Topic == "" {
		return errors.NewValidationErrorf("topic is required")
	}
	if opts.GrainTimeout <= 0 {
		return errors.NewValidationErrorf("grain timeout must be greater than zero")
	}
	if opts.RetrierFactory == nil {
		return errors.NewValidationErrorf("retrier factory is required")
	}
	return nil
}

type GrainSystem struct {
	si                   *silo.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	grainPersistence     ports.GrainStorage
	eventPub             pkgPorts.EventPublisher
	logger               pkgPorts.Logger
	topic                contracts.Topic
	grainTimeout         time.Duration
}

func NewGrainSystem(opts *GrainSystemOptions) (*GrainSystem, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	si, err := silo.NewSilo(&silo.SiloOptions{
		Timeout:        opts.GrainTimeout,
		Logger:         opts.Logger,
		RetrierFactory: opts.RetrierFactory,
	})
	if err != nil {
		return nil, err
	}

	grainOpts := &grains.RideGrainOptions{
		Storage:  opts.Storage,
		EventPub: opts.EventPublisher,
		Logger:   opts.Logger,
		Topic:    opts.Topic,
	}
	if err := grainOpts.Validate(); err != nil {
		return nil, err
	}
	si.RegisterFactory(domain.RideGrainKind, func(identity *grain.GrainIdentity) pkgPorts.Grain {
		return grains.NewRideGrain(grainOpts)
	})

	return &GrainSystem{
		si:                   si,
		eventPub:             opts.EventPublisher,
		logger:               opts.Logger,
		topic:                opts.Topic,
		grainPersistence:     opts.Storage,
		grainIdentityFactory: &service.GrainIdentityFactory{},
		grainTimeout:         opts.GrainTimeout,
	}, nil
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
