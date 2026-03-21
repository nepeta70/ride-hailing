package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	pkgMongo "github.com/nepeta70/ride-hailing/internal/pkg/adapters/mongodb"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/driver/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/driver/internal/adapters/mongodb"
	"github.com/nepeta70/ride-hailing/services/driver/internal/config"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	tel, err := telemetry.NewTelemetryProvider(ctx, &cfg.BaseConfig)
	if err != nil {
		log.Printf("ERROR: Failed to create telemetry provider: %v", err)
		return
	}
	defer tel.Shutdown(ctx)

	logger := tel.Logger()

	retrierFactory := retry.NewRetrierFactory(tel)

	mongoAdapter, err := pkgMongo.NewMongoAdapter(&pkgMongo.MongoAdapterOpts{
		Config:         &cfg.Mongo,
		RetrierFactory: retrierFactory,
		Telemetry:      tel,
	}, ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create mongo adapter:", "error", err)
		return
	}
	defer mongoAdapter.Close()

	repo, err := mongodb.NewDriverRepository(ctx, cfg, mongoAdapter, tel)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create driver repository:", "error", err)
		return
	}

	contextManager := ctxmgr.NewContextManager()
	driverService := service.NewDriverService(repo, tel)
	handler := grpcAdapters.NewDriverHandler(driverService, tel, contextManager)

	opts := &grpc_adapter.GRPGAdapterOptions{
		Config:            &cfg.BaseConfig,
		Logger:            logger,
		ContextManager:    contextManager,
		AuthConfiguration: grpcAdapters.NewEndpointRoles(&cfg.BaseConfig),
		Telemetry:         tel,
	}
	grpcServer, err := grpc_adapter.NewGRPCAdapter(opts)
	if err != nil {
		logger.Error("Failed to create gRPC adapter:", "error", err)
		return
	}
	grpcServer.RegisterService(&driverv1.DriverService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, mongoAdapter)

	grpcServer.Run(ctx)
}
