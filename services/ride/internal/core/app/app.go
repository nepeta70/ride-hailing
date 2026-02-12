package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type Application struct {
	Commands           *Commands
	Queries            *Queries
	Logger             pkgPorts.Logger
	ContextManager     *ctxmgr.ContextManager
	storage            ports.StorageBundle
	grainSystem        *GrainSystem
	distanceCalculator service.DirectionsEstimator
	directionsService  ports.DirectionsService
	config             *config.Config
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

func NewApplication(cfg *config.Config, logger pkgPorts.Logger, distanceCalculator *service.DirectionsEstimator, directionsService ports.DirectionsService, storage ports.StorageBundle, grainSystem *GrainSystem, contextManager *ctxmgr.ContextManager) *Application {
	app := &Application{
		Commands: &Commands{
			EstimateFare:       commands.NewEstimateFareHandler(cfg, logger, storage, distanceCalculator, directionsService),
			CreateFareRate:     commands.NewCreateFareRateHandler(storage.FareRatesWriteRepo()),
			UpdateFareRate:     commands.NewUpdateFareRateHandler(storage.FareRatesWriteRepo()),
			RequestRide:        commands.NewRequestRideHandler(cfg, storage, grainSystem, logger),
			CancelRide:         commands.NewCancelRideHandler(cfg, storage, grainSystem, logger),
			AcceptOrRejectRide: commands.NewAcceptOrRejectRideHandler(cfg, storage, grainSystem, logger),
			StartRide:          commands.NewStartRideHandler(cfg, storage, grainSystem, logger),
			CompleteRide:       commands.NewCompleteRideHandler(cfg, storage, grainSystem, logger),
		},
		Queries: &Queries{
			FareRates: queries.NewGetFareRatesHandler(nil),
		},
		Logger:             logger,
		ContextManager:     contextManager,
		config:             cfg,
		storage:            storage,
		distanceCalculator: *distanceCalculator,
		directionsService:  directionsService,
		grainSystem:        grainSystem,
	}
	return app
}
