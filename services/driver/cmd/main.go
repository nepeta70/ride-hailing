package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
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

	logger := cfg.Logging.ConfigureLogger()

	driverService := service.NewDriverService()
	handler := grpcAdapters.NewDriverHandler(driverService)

	grpcServer := grpc_adapter.NewGRPCAdapter("Driver Service", &cfg.BaseConfig, logger)
	grpcServer.RegisterService(&driverv1.DriverService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
