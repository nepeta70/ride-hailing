package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
)

type ApplicationOpts struct {
	Config         *config.Config
	ContextManager *ctxmgr.ContextManager
	Telemetry      pkgPorts.TelemetryProvider
	LocationRepo   ports.LocationRepository
	RetryFactory   pkgPorts.RetrierFactoryInterface
}

func (opts *ApplicationOpts) Validate() error {
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager cannot be nil")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry cannot be nil")
	}
	if opts.LocationRepo == nil {
		return errors.NewValidationErrorf("location repository cannot be nil")
	}
	if opts.RetryFactory == nil {
		return errors.NewValidationErrorf("retry factory cannot be nil")
	}
	return nil
}

type Application struct {
	ContextManager  *ctxmgr.ContextManager
	Telemetry       pkgPorts.TelemetryProvider
	LocationService *service.LocationService
}

func NewApplication(opts *ApplicationOpts) (*Application, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	locationService := service.NewLocationService(&service.LocationServiceOpts{
		Config:    opts.Config,
		Repo:      opts.LocationRepo,
		Telemetry: opts.Telemetry,
	})
	return &Application{
		ContextManager:  opts.ContextManager,
		Telemetry:       opts.Telemetry,
		LocationService: locationService,
	}, nil
}
