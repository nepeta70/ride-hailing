package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	notificationv1 "github.com/nepeta70/ride-hailing/gen/proto/notification/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
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

	logger := cfg.Logging.ConfigureLogger()

	handler := grpc.NewNotificationHandler()
	grpcServer := grpc_adapter.NewGRPCAdapter("Notification Service", &cfg.BaseConfig, logger)
	grpcServer.RegisterService(&notificationv1.NotificationService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
