package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/matching/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/matching/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	matchingService := service.NewMatchingService()
	handler := grpcAdapters.NewMatchingHandler(matchingService)

	tel, err := telemetry.NewTelemetryProvider(ctx, cfg.ServiceName, &cfg.Telemetry, logger)
	if err != nil {
		logger.Error("Failed to create telemetry provider:", "error", err)
		return
	}
	defer tel.Shutdown(ctx)

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
	grpcServer.RegisterService(&matchingv1.MatchingService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
