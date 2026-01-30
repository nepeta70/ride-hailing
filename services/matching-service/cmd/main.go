package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/matching-service/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/matching-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching-service/internal/core/service"
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

	matchingService := service.NewMatchingService()
	handler := grpcAdapters.NewMatchingHandler(matchingService)

	grpcServer := grpc_adapter.NewGRPCAdapter("Matching Service", &cfg.BaseConfig, logger)
	matchingv1.RegisterMatchingServiceServer(grpcServer.Server, handler)
	reflection.Register(grpcServer.Server)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
