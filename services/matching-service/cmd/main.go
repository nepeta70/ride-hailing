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

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/matching-service/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/matching-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching-service/internal/core/service"
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
	// In your main.go setup:
	const maxMsgSize = 1024 * 1024 * 2 // 2MB

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		// Chain your other interceptors here
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerLogging(logger),
			middleware.UnaryRateLimit(100, 10),
			middleware.UnaryServerLogging(logger),
		),
	)

	matchingService := service.NewMatchingService()
	handler := grpcAdapters.NewMatchingHandler(matchingService)

	matchingv1.RegisterMatchingServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// 3. Register Standard Health Service
	healthService := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthService)

	go func() {
		for {
			status := healthpb.HealthCheckResponse_SERVING
			healthService.SetServingStatus("order-service", status)
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
