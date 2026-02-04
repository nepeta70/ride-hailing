package grpc_adapter

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCAdapter struct {
	Server        *grpc.Server
	Listener      net.Listener
	HealthService *health.Server
	Address       string
	logger        ports.Logger
	config        *config.BaseConfig
	serviceName   string
}

func NewGRPCAdapter(serviceName string, cfg *config.BaseConfig, logger ports.Logger) GRPCAdapter {
	grpcAddr := ":" + strconv.Itoa(cfg.Server.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to listen: %v", "error", err)
	}
	maxMsgSize := int(cfg.Security.MaxBodyBytes)

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.ChainUnaryInterceptor(
			middleware.UnaryTimeout(cfg.Server.WriteTimeout),
			middleware.UnaryRateLimit(rate.Limit(cfg.Security.RateLimit), cfg.Security.RateBurst),
			middleware.UnaryServerLogging(logger),
		),
	)

	healthService := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthService)

	return GRPCAdapter{
		Server:        grpcServer,
		Listener:      lis,
		HealthService: healthService,
		Address:       grpcAddr,
		logger:        logger,
		config:        cfg,
		serviceName:   serviceName,
	}
}

func (s *GRPCAdapter) MonitorHealth(ctx context.Context, providers ...ports.HealthProvider) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, p := range providers {
					status := healthpb.HealthCheckResponse_SERVING
					if err := p.HealthCheck(ctx); err != nil {
						s.logger.Warn("Health check failed", "service", p.ServiceName(), "error", err)
						status = healthpb.HealthCheckResponse_NOT_SERVING
					}
					s.HealthService.SetServingStatus(p.ServiceName(), status)
				}
			}
		}
	}()
}

func (s *GRPCAdapter) Run(ctx context.Context) {
	s.logger.Info("Running " + s.serviceName + " gRPC server")
	// 1. Start serving in background
	go func() {
		s.logger.Info("gRPC server starting", "addr", s.Address)
		if err := s.Server.Serve(s.Listener); err != nil && err != grpc.ErrServerStopped {
			s.logger.Error("gRPC server error", "error", err)
		}
	}()

	// 2. Wait for context cancellation (SIGTERM/Interrupt)
	<-ctx.Done()
	s.logger.Info("shutting down " + s.serviceName + " gRPC server	...")

	// 3. Graceful Shutdown logic
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.Timeouts.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Graceful shutdown of " + s.serviceName + " gRPC server complete")
	case <-shutdownCtx.Done():
		s.logger.Warn("Graceful shutdown timed out, forcing stop")
		s.Server.Stop()
	}

	s.logger.Info("shutdown " + s.serviceName + " gRPC server complete")
}

func (s *GRPCAdapter) RegisterService(server *grpc.ServiceDesc, handler any) {
	s.Server.RegisterService(server, handler)
	reflection.Register(s.Server)
}
