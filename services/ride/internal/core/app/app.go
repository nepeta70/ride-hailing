package app

import (
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
	storage            ridePorts.StorageBundle
	distanceCalculator service.DirectionsEstimator
	directionsService  ridePorts.DirectionsService
	logger             ports.Logger
	config             *config.Config
}

type Commands struct {
	CreateFareRate     *commands.CreateFareRateHandler
	UpdateFareRate     *commands.UpdateFareRateHandler
	RequestRide        *commands.RequestRideHandler
	CancelRide         *commands.CancelRideHandler
	AcceptOrRejectRide *commands.AcceptOrRejectRideHandler
	StartRide          *commands.StartRideHandler
	CompleteRide       *commands.CompleteRideHandler
}

type Queries struct {
	FareEstimate *queries.FareEstimateHandler
	FareRates    *queries.GetFareRateHandler
}

func NewApplication(cfg *config.Config, logger ports.Logger, distanceCalculator *service.DirectionsEstimator, directionsService ridePorts.DirectionsService, storage ridePorts.StorageBundle) *Application {
	app := &Application{
		Commands: &Commands{
			CreateFareRate:     commands.NewCreateFareRateHandler(storage.FareRatesWriteRepo()),
			UpdateFareRate:     commands.NewUpdateFareRateHandler(storage.FareRatesWriteRepo()),
			RequestRide:        commands.NewRequestRideHandler(storage.RideWriteRepo()),
			CancelRide:         commands.NewCancelRideHandler(storage.RideWriteRepo()),
			AcceptOrRejectRide: commands.NewAcceptOrRejectRideHandler(storage.RideWriteRepo()),
			StartRide:          commands.NewStartRideHandler(storage.RideWriteRepo()),
			CompleteRide:       commands.NewCompleteRideHandler(storage.RideWriteRepo()),
		},
		Queries: &Queries{
			FareEstimate: queries.NewFareEstimateHandler(cfg, distanceCalculator, directionsService, storage.FareRatesReadRepo()),
			FareRates:    queries.NewGetFareRatesHandler(nil),
		},
		logger:  logger,
		config:  cfg,
		storage: storage,
	}
	return app
}
