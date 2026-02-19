package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
)

type Application struct {
	ContextManager  *ctxmgr.ContextManager
	Logger          pkgPorts.Logger
	LocationService *service.LocationService
}

func NewApplication(
	ctxMgr *ctxmgr.ContextManager,
	logger pkgPorts.Logger,
	locationRepo ports.LocationRepository,
) *Application {
	locationService := service.NewLocationService(locationRepo)
	return &Application{
		ContextManager:  ctxMgr,
		Logger:          logger,
		LocationService: locationService,
	}
}
