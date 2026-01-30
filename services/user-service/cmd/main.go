package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/user-service/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app"
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

	application := app.NewApplication(cfg, logger)
	handler := grpcAdapters.NewUserHandler(application)
	grpcServer := grpc_adapter.NewGRPCAdapter("User Service", &cfg.BaseConfig, logger)
	userv1.RegisterUserServiceServer(grpcServer.Server, handler)
	reflection.Register(grpcServer.Server)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
