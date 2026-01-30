package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/redis"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/location-service/internal/adapters/grpc"
	redisAdapters "github.com/nepeta70/ride-hailing/services/location-service/internal/adapters/redis"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/service"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Error("failed to init redis: %v", "error", err)
	}
	defer redisClient.Close()

	locationRepository := redisAdapters.NewRedisRepository(cfg, redisClient, logger)
	locationService := service.NewLocationService(locationRepository)
	handler := grpcAdapters.NewLocationHandler(locationService)

	grpcServer := grpc_adapter.NewGRPCAdapter("Location Service", &cfg.BaseConfig, logger)
	locationv1.RegisterLocationServiceServer(grpcServer.Server, handler)
	reflection.Register(grpcServer.Server)

	grpcServer.MonitorHealth(ctx, redisClient)

	grpcServer.Run(ctx)
}
