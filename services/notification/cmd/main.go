package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	notificationv1 "github.com/nepeta70/ride-hailing/gen/proto/notification/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/services/notification/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/notification/internal/config"
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

	logger := tel.GetLogger()

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
	handler := grpc.NewNotificationHandler()
	grpcServer.RegisterService(&notificationv1.NotificationService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
