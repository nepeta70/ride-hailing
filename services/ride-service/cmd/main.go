package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/adapters/googlemaps"
	grpcHandler "github.com/nepeta70/ride-hailing/services/ride-service/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	fareService := service.NewFareService( /* dependencies */ )
	rideService := service.NewRideService( /* dependencies */ )

	directionsEstimator := service.NewDirectionsEstimator()
	googleMaps, err := googlemaps.NewGoogleMapsAdapter(cfg)
	if err != nil {
		logger.Error("Failed to create Google Maps adapter: %v", "GoogleMapsAdapter", err)
	}
	application := app.NewApplication(cfg, logger, directionsEstimator, googleMaps)

	handler := grpcHandler.NewRideHandler(application, fareService, rideService)

	grpcServer := grpc_adapter.NewGRPCAdapter("Ride Service", &cfg.BaseConfig, logger)
	grpcServer.RegisterService(&ridev1.RideService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
