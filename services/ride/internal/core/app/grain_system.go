package app

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/silo"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
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
	Topic          contracts.Topic
	GrainTimeout   time.Duration
	RetrierFactory *retry.RetrierFactory
	ContextManager *ctxmgr.ContextManager
	Telemetry      pkgPorts.TelemetryProvider
}

func (opts *GrainSystemOptions) Validate() error {
	if opts.Storage == nil {
		return errors.NewValidationErrorf("grain storage is required")
	}
	if opts.EventPublisher == nil {
		return errors.NewValidationErrorf("event publisher is required")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
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
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	return nil
}

type GrainSystem struct {
	si                   *silo.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	grainPersistence     ports.GrainStorage
	eventPub             pkgPorts.EventPublisher
	telemetry            pkgPorts.TelemetryProvider
	topic                contracts.Topic
	grainTimeout         time.Duration
	contextManager       *ctxmgr.ContextManager
}

func NewGrainSystem(opts *GrainSystemOptions) (*GrainSystem, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	si, err := silo.NewSilo(&silo.SiloOptions{
		Timeout:        opts.GrainTimeout,
		Telemetry:      opts.Telemetry,
		RetrierFactory: opts.RetrierFactory,
	})
	if err != nil {
		return nil, err
	}

	grainOpts := &grains.RideGrainOptions{
		Storage:        opts.Storage,
		EventPub:       opts.EventPublisher,
		Telemetry:      opts.Telemetry,
		Topic:          opts.Topic,
		ContextManager: opts.ContextManager,
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
		telemetry:            opts.Telemetry,
		topic:                opts.Topic,
		grainPersistence:     opts.Storage,
		grainIdentityFactory: &service.GrainIdentityFactory{},
		grainTimeout:         opts.GrainTimeout,
		contextManager:       opts.ContextManager,
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
