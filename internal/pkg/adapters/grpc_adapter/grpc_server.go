package grpc_adapter

import (
	"context"
	"fmt"

	"net"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPGAdapterOptions struct {
	ServiceName            string
	Config                 *config.BaseConfig
	Logger                 ports.Logger
	ContextManager         *ctxmgr.ContextManager
	AuthConfiguration      ports.EndpointRoles
	AdditionalInterceptors []grpc.UnaryServerInterceptor
}

func (opts *GRPGAdapterOptions) Validate() error {
	if opts.ServiceName == "" {
		return errors.NewValidationErrorf("service name is required")
	}
	if opts.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if opts.AuthConfiguration == nil {
		return errors.NewValidationErrorf("auth configuration is required")
	}
	return nil
}

type GRPCAdapter struct {
	Server        *grpc.Server
	Listener      net.Listener
	HealthService *health.Server
	Address       string
	logger        ports.Logger
	config        *config.BaseConfig
	serviceName   string
}

func NewGRPCAdapter(opts *GRPGAdapterOptions) (*GRPCAdapter, error) {
	if err := opts.Validate(); err != nil {
		opts.Logger.Error("FATAL: failed to create GRPC adapter", "error", err)
		return nil, err
	}
	grpcAddr := fmt.Sprintf(":%d", opts.Config.Server.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		opts.Logger.Error("failed to listen: %v", "error", err)
		return nil, err
	}
	reg := prometheus.NewRegistry()
	filteredChainOpts := &middleware.FilteredChainOpts{
		Config:                 opts.Config,
		Logger:                 opts.Logger,
		ContextManager:         opts.ContextManager,
		EndpointRoles:          opts.AuthConfiguration,
		AdditionalInterceptors: opts.AdditionalInterceptors,
		Metrics:                telemetry.NewMetrics("ride-hailing", opts.ServiceName, reg),
	}
	filteredChain, err := middleware.NewInterceptorChain(filteredChainOpts)
	if err != nil {
		opts.Logger.Error("CRITICAL: failed to create interceptor chain", "error", err)
		return nil, err
	}

	maxMsgSize := int(opts.Config.Security.MaxBodyBytes)
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.ChainUnaryInterceptor(filteredChain.FilteredChain()),
	)

	healthService := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthService)

	return &GRPCAdapter{
		Server:        grpcServer,
		Listener:      lis,
		HealthService: healthService,
		Address:       grpcAddr,
		logger:        opts.Logger,
		config:        opts.Config,
		serviceName:   opts.ServiceName,
	}, nil
}

func (s *GRPCAdapter) MonitorHealth(ctx context.Context, providers ...ports.HealthProvider) {
	ticker := time.NewTicker(10 * time.Second) // TODO: move to config

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
