package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/user/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/user/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/user/internal/config"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app"
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

	repo := inmemory.NewInMemoryUserRepository()
	application := app.NewApplication(cfg, logger, repo, repo)
	handler := grpcAdapters.NewUserHandler(application)

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
	grpcServer.RegisterService(&userv1.UserService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
