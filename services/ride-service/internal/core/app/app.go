package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/service"
	ridePorts "github.com/nepeta70/ride-hailing/services/ride-service/internal/ports"
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
}

func newCommands() *Commands {
	return &Commands{}
}

type Queries struct {
	GetFare queries.GetFareHandler
}

func newQueries(config *config.Config, distanceCalculator *service.DirectionsEstimator, directionsService ridePorts.DirectionsService) *Queries {
	return &Queries{
		GetFare: queries.NewGetFareHandler(config, distanceCalculator, directionsService),
	}
}

func NewApplication(cfg *config.Config, logger ports.Logger, distanceCalculator *service.DirectionsEstimator, directionsService ridePorts.DirectionsService, storage ridePorts.StorageBundle) *Application {
	app := &Application{
		Commands: newCommands(),
		Queries:  newQueries(cfg, distanceCalculator, directionsService),
		logger:   logger,
		config:   cfg,
		storage:  storage,
	}
	return app
}
