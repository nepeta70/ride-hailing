package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/location/internal/adapters/grpc"
	rdstore "github.com/nepeta70/ride-hailing/services/location/internal/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/app"
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
		logger.Error("Failed to init Redis:", "error", err)
		return
	}
	defer redisClient.Close()

	locationRepository := rdstore.NewRedisRepository(cfg, redisClient, logger)
	app := app.NewApplication(ctxmgr.NewContextManager(), logger, locationRepository)

	handler := grpcAdapters.NewLocationHandler(app)

	tel, err := telemetry.NewTelemetryProvider(ctx, cfg.ServiceName, &cfg.Telemetry, logger)
	if err != nil {
		logger.Error("Failed to create telemetry provider:", "error", err)
		return
	}
	defer tel.Shutdown(ctx)

	opts := &grpc_adapter.GRPGAdapterOptions{
		Config:            &cfg.BaseConfig,
		Logger:            logger,
		ContextManager:    ctxmgr.NewContextManager(),
		AuthConfiguration: grpcAdapters.NewEndpointRoles(&cfg.BaseConfig),
		Telemetry:         tel,
	}
	grpcServer, err := grpc_adapter.NewGRPCAdapter(opts)
	if err != nil {
		logger.Error("Failed to create gRPC adapter:", "error", err)
		return
	}
	grpcServer.RegisterService(&locationv1.LocationService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, redisClient)

	grpcServer.Run(ctx)
}
