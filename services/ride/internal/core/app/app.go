package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	ridePorts "github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type Application struct {
	Commands           *Commands
	Queries            *Queries
	Logger             ports.Logger
	ContextManager     *ctxmgr.ContextManager
	storage            ridePorts.StorageBundle
	distanceCalculator service.DirectionsEstimator
	directionsService  ridePorts.DirectionsService
	config             *config.Config
}

type Commands struct {
	FareEstimate       *commands.FareEstimateHandler
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

func NewApplication(cfg *config.Config, logger ports.Logger, distanceCalculator *service.DirectionsEstimator, directionsService ridePorts.DirectionsService, storage ridePorts.StorageBundle, contextManager *ctxmgr.ContextManager) *Application {
	app := &Application{
		Commands: &Commands{
			FareEstimate:       commands.NewEstimateFareHandler(cfg, logger, storage, distanceCalculator, directionsService),
			CreateFareRate:     commands.NewCreateFareRateHandler(storage.FareRatesWriteRepo()),
			UpdateFareRate:     commands.NewUpdateFareRateHandler(storage.FareRatesWriteRepo()),
			RequestRide:        commands.NewRequestRideHandler(cfg, storage, logger),
			CancelRide:         commands.NewCancelRideHandler(storage.RideWriteRepo()),
			AcceptOrRejectRide: commands.NewAcceptOrRejectRideHandler(storage.RideWriteRepo()),
			StartRide:          commands.NewStartRideHandler(storage.RideWriteRepo()),
			CompleteRide:       commands.NewCompleteRideHandler(storage.RideWriteRepo()),
		},
		Queries: &Queries{
			FareRates: queries.NewGetFareRatesHandler(nil),
		},
		Logger:         logger,
		ContextManager: contextManager,
		config:         cfg,
		storage:        storage,
	}
	return app
}
