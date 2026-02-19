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
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"

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

	tel, err := telemetry.NewTelemetryProvider(ctx, cfg.ServiceName, &cfg.Telemetry, logger)
	if err != nil {
		logger.Error("Failed to create telemetry provider:", "error", err)
		return
	}
	defer tel.Shutdown(ctx)

	retrierFactory := retry.NewRetrierFactory(logger, tel.GetMetrics())

	redisClient, err := rd.NewClient(&cfg.Redis, retrierFactory, logger)
	if err != nil {
		logger.Error("Failed to init Redis:", "error", err)
		return
	}
	defer redisClient.Close()

	pg, err := pgstore.NewPostgresDB(&cfg.Postgres, retrierFactory, logger)
	if err != nil {
		logger.Error("Failed to create Postgres DB:", "error", err)
		return
	}

	topicProvider := service.NewTopicProvider()

	eventPublisher := pubsub.NewEventPublisher(cfg.Kafka, topicProvider, logger)
	defer eventPublisher.Close()

	storage, err := adapters.NewRedisStorageBundle(&adapters.StorageBundleOptions{
		Config:   cfg,
		RdClient: redisClient,
		PgDB:     pg,
		Logger:   logger,
	})
	if err != nil {
		logger.Error("Failed to create storage adapter:", "error", err)
		return
	}
	defer storage.Close()

	googleMaps, err := googlemaps.NewGoogleMapsAdapter(&googlemaps.GoogleMapsAdapterOptions{
		APIKey:          cfg.KeysConfig.GoogleMapsAPIKey,
		FallBackService: service.NewDirectionsEstimator(),
		Logger:          logger,
	})
	if err != nil {
		logger.Error("Failed to create Google Maps adapter:", "error", err)
		googleMaps = nil // Set to nil to allow service to degrade gracefully
	}

	grainSystem, err := app.NewGrainSystem(&app.GrainSystemOptions{
		Topic:          contracts.TopicRide,
		GrainTimeout:   cfg.Timeouts.RequestTimeout,
		Storage:        storage.GrainStorage(),
		EventPublisher: eventPublisher,
		Logger:         logger,
		RetrierFactory: retrierFactory,
	})
	if err != nil {
		logger.Error("Failed to create grain system:", "error", err)
		return
	}
	application, err := app.NewApplication(&app.ApplicationOptions{
		Config:            cfg,
		Logger:            logger,
		DirectionsService: googleMaps,
		Storage:           storage,
		GrainSystem:       grainSystem,
		ContextManager:    ctxmgr.NewContextManager(),
	})
	if err != nil {
		logger.Error("Failed to create application:", "error", err)
		return
	}

	handler := grpcHandler.NewRideHandler(application, storage, grainSystem)

	opts := &grpc_adapter.GRPGAdapterOptions{
		Config:            &cfg.BaseConfig,
		Logger:            logger,
		ContextManager:    ctxmgr.NewContextManager(),
		AuthConfiguration: grpcHandler.NewEndpointRoles(&cfg.BaseConfig),
		Telemetry:         tel,
	}
	grpcServer, err := grpc_adapter.NewGRPCAdapter(opts)
	if err != nil {
		logger.Error("Failed to create gRPC adapter:", "error", err)
		return
	}
	grpcServer.RegisterService(&ridev1.RideService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, redisClient, pg)

	grpcServer.Run(ctx)
}
