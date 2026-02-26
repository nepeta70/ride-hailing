package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/driver/internal/adapters/grpc"
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

	driverService := service.NewDriverService()
	handler := grpcAdapters.NewDriverHandler(driverService)

	opts := &grpc_adapter.GRPGAdapterOptions{
		Config:         &cfg.BaseConfig,
		Logger:         logger,
		ContextManager: ctxmgr.NewContextManager(),
		//AuthConfiguration: grpcAdapters.NewEndpointRoles(&cfg.BaseConfig), TODO: implemnt it
		Telemetry: tel,
	}
	grpcServer, err := grpc_adapter.NewGRPCAdapter(opts)
	if err != nil {
		logger.Error("Failed to create gRPC adapter:", "error", err)
		return
	}
	grpcServer.RegisterService(&driverv1.DriverService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
