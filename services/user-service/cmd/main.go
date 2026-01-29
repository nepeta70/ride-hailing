package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/user-service/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	logger.Debug("Running Application")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	grpcAddr := ":" + strconv.Itoa(cfg.Server.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to listen: %v", "error", err)
	}
	grpcServer := grpc.NewServer()

	application := app.NewApplication(cfg, logger)
	handler := grpcAdapters.NewUserHandler(application)

	userv1.RegisterUserServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// 3. Register Standard Health Service
	healthService := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthService)

	go func() {
		for {
			status := healthpb.HealthCheckResponse_SERVING
			healthService.SetServingStatus("user-service", status)
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		logger.Info("gRPC server starting", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	logger.Info("server listening", "addr", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
	}

	<-ctx.Done()
	logger.Debug("shutting down...")

	grpcServer.GracefulStop()

	logger.Debug("shutdown complete")
}
