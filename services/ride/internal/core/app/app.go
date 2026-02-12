package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type ApplicationOptions struct {
	Config            *config.Config
	Logger            pkgPorts.Logger
	DirectionsService ports.DirectionsService
	Storage           ports.StorageBundle
	GrainSystem       *GrainSystem
	ContextManager    *ctxmgr.ContextManager
}

type Application struct {
	Commands          *Commands
	Queries           *Queries
	Logger            pkgPorts.Logger
	ContextManager    *ctxmgr.ContextManager
	storage           ports.StorageBundle
	grainSystem       *GrainSystem
	directionsService ports.DirectionsService
	config            *config.Config
}

type Commands struct {
	EstimateFare       *commands.EstimateFareHandler
	CreateFareRate     *commands.CreateFareRateHandler
	UpdateFareRate     *commands.UpdateFareRateHandler
	RequestRide        *commands.RequestRideHandler
	CancelRide         *commands.CancelRideHandler
	AcceptOrRejectRide *commands.AcceptOrRejectRideHandler
	StartRide          *commands.StartRideHandler
	CompleteRide       *commands.CompleteRideHandler
}

type Queries struct {
	FareRates *queries.GetFareRateHandler
}

func NewApplication(opts *ApplicationOptions) *Application {
	app := &Application{
		Commands: &Commands{
			EstimateFare:       commands.NewEstimateFareHandler(opts.Config, opts.Logger, opts.Storage, opts.DirectionsService),
			CreateFareRate:     commands.NewCreateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			UpdateFareRate:     commands.NewUpdateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			RequestRide:        commands.NewRequestRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Logger),
			CancelRide:         commands.NewCancelRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Logger),
			AcceptOrRejectRide: commands.NewAcceptOrRejectRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Logger),
			StartRide:          commands.NewStartRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Logger),
			CompleteRide:       commands.NewCompleteRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Logger),
		},
		Queries: &Queries{
			FareRates: queries.NewGetFareRatesHandler(nil),
		},
		Logger:            opts.Logger,
		ContextManager:    opts.ContextManager,
		config:            opts.Config,
		storage:           opts.Storage,
		directionsService: opts.DirectionsService,
		grainSystem:       opts.GrainSystem,
	}
	return app
}
