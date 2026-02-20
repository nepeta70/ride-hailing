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
	Logger         pkgPorts.Logger
	Metrics        pkgPorts.Metrics
	LocationRepo   ports.LocationRepository
	RetryFactory   pkgPorts.RetrierFactoryInterface
}

func (opts *ApplicationOpts) Validate() error {
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager cannot be nil")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger cannot be nil")
	}
	if opts.LocationRepo == nil {
		return errors.NewValidationErrorf("location repository cannot be nil")
	}
	if opts.RetryFactory == nil {
		return errors.NewValidationErrorf("retry factory cannot be nil")
	}
	if opts.Metrics == nil {
		return errors.NewValidationErrorf("metrics cannot be nil")
	}
	return nil
}

type Application struct {
	ContextManager  *ctxmgr.ContextManager
	Logger          pkgPorts.Logger
	LocationService *service.LocationService
}

func NewApplication(opts *ApplicationOpts) (*Application, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	locationService := service.NewLocationService(&service.LocationServiceOpts{
		Config:  opts.Config,
		Repo:    opts.LocationRepo,
		Logger:  opts.Logger,
		Metrics: opts.Metrics,
	})
	return &Application{
		ContextManager:  opts.ContextManager,
		Logger:          opts.Logger,
		LocationService: locationService,
	}, nil
}
