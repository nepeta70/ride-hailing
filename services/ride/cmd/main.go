package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"

	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/googlemaps"
	grpcHandler "github.com/nepeta70/ride-hailing/services/ride/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	redisClient, err := rd.NewClient(&cfg.Redis, logger)
	if err != nil {
		logger.Error("failed to init redis: %v", "error", err)
	}
	defer redisClient.Close()

	pg, err := pgstore.NewPostgresDB(&cfg.Postgres, logger)
	if err != nil {
		logger.Error("Failed to create Postgres DB: %v", "PostgresDB", err)
		return
	}

	storage, err := adapters.NewRedisStorageBundle(cfg, redisClient, pg, logger)
	if err != nil {
		logger.Error("Failed to create storage adapter: %v", "StorageAdapter", err)
		return
	}
	defer storage.Close()

	directionsEstimator := service.NewDirectionsEstimator()
	googleMaps, err := googlemaps.NewGoogleMapsAdapter(cfg)
	if err != nil {
		logger.Error("Failed to create Google Maps adapter: %v", "GoogleMapsAdapter", err)
		googleMaps = nil // Set to nil to allow service to degrade gracefully
	}

	contextManager := ctxmgr.NewContextManager()
	grainSystem := app.NewGrainSystem(&cfg.BaseConfig, storage.GrainStorage(), nil, nil, logger) // Pass nil for event publisher for now
	application := app.NewApplication(cfg, logger, directionsEstimator, googleMaps, storage, grainSystem, contextManager)

	handler := grpcHandler.NewRideHandler(application, storage, grainSystem)

	grpcServer := grpc_adapter.NewGRPCAdapter("Ride Service", &cfg.BaseConfig, logger)
	grpcServer.RegisterService(&ridev1.RideService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, redisClient, pg, googleMaps)

	grpcServer.Run(ctx)
}
