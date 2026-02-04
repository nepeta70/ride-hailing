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
	grpcAdapters "github.com/nepeta70/ride-hailing/services/location/internal/adapters/grpc"
	rdstore "github.com/nepeta70/ride-hailing/services/location/internal/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	redisClient, err := rd.NewClient(&cfg.Redis)
	if err != nil {
		logger.Error("failed to init redis: %v", "error", err)
	}
	defer redisClient.Close()

	locationRepository := rdstore.NewRedisRepository(cfg, redisClient, logger)
	locationService := service.NewLocationService(locationRepository)
	handler := grpcAdapters.NewLocationHandler(locationService)

	grpcServer := grpc_adapter.NewGRPCAdapter("Location Service", &cfg.BaseConfig, logger)
	grpcServer.RegisterService(&locationv1.LocationService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, redisClient)

	grpcServer.Run(ctx)
}
