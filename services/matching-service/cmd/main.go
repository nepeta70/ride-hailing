package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nepeta70/ride-hailing/services/matching-service/internal/config"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	logger := cfg.Logging.ConfigureLogger()

	logger.Debug("Running Application")

	// // 2. Setup gRPC Server
	// grpcAddr := ":" + strconv.Itoa(cfg.Server.Port)
	// lis, err := net.Listen("tcp", grpcAddr)
	// if err != nil {
	// 	logger.Error("failed to listen: %v", "error", err)
	// }
	// grpcServer := grpc.NewServer()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// matchingService := service.NewMatchingService( /* dependencies */ )
	// handler := grpcHandler.NewMatchingHandler(matchingService)

	// matchingv1.RegisterMatchingServiceServer(grpcServer, handler)
	// reflection.Register(grpcServer)

	// // 3. Register Standard Health Service
	healthService := health.NewServer()
	// healthpb.RegisterHealthServer(grpcServer, healthService)

	// 4. Run a background goroutine to update status based on Redis health
	go func() {
		for {
			status := healthpb.HealthCheckResponse_SERVING
			// if err := redisClient.HealthCheck(context.Background()); err != nil {
			// 	logger.Warn("Redis health check failed", "error", err)
			// 	status = healthpb.HealthCheckResponse_NOT_SERVING
			// }

			// Update the status for the whole service or a specific component
			healthService.SetServingStatus("matching-service", status)

			time.Sleep(10 * time.Second)
		}
	}()

	// go func() {
	// 	logger.Info("gRPC server starting", "addr", grpcAddr)
	// 	if err := grpcServer.Serve(lis); err != nil {
	// 		logger.Error("gRPC server error", "error", err)
	// 	}
	// }()

	// logger.Info("server listening", "addr", lis.Addr())
	// if err := grpcServer.Serve(lis); err != nil {
	// 	logger.Error("failed to serve", "error", err)
	// }

	<-ctx.Done()
	logger.Debug("shutting down...")

	//grpcServer.GracefulStop()

	logger.Debug("shutdown complete")
}
