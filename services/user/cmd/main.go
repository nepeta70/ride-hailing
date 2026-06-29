package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	pg "github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/user/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/user/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/user/internal/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/user/internal/adapters/security"
	"github.com/nepeta70/ride-hailing/services/user/internal/config"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/validator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Server.Port)
	}

	tel, err := telemetry.NewTelemetryProvider(ctx, &cfg.BaseConfig)
	if err != nil {
		log.Printf("ERROR: Failed to create telemetry provider: %v", err)
		return
	}
	defer tel.Shutdown(ctx)

	logger := tel.Logger()

	retrierFactory := retry.NewRetrierFactory(tel)
	contextManager := ctxmgr.NewContextManager()

	postgres, err := pg.NewPostgresDB(&pg.PostgresOpts{
		Config:         &cfg.Postgres,
		RetrierFactory: retrierFactory,
		Telemetry:      tel,
		ContextManager: contextManager,
	})
	if err != nil {
		logger.Error("Failed to create Postgres DB:", "error", err)
		return
	}

	repo := inmemory.NewInMemoryUserRepository() // TODO replace it by pgstore implementation

	appOpts := &app.AppOpts{
		Config:      cfg,
		Telemetry:   tel,
		WriteRepo:   repo,
		ReadRepo:    repo,
		Hasher:      security.NewBcryptHasher(),
		Credentials: pgstore.NewCredentialsRepo(cfg, postgres),
		Validator:   validator.NewPasswordValidator(),
	}
	application, err := app.NewApplication(appOpts)
	if err != nil {
		logger.Error("Failed to create Application:", "error", err)
		return
	}
	handler := grpcAdapters.NewUserHandler(application)

	opts := &grpc_adapter.GRPGAdapterOptions{
		Config:            &cfg.BaseConfig,
		Logger:            logger,
		ContextManager:    ctxmgr.NewContextManager(),
		AuthConfiguration: grpcAdapters.NewEndpointRoles(&cfg.BaseConfig),
		Telemetry:         tel,
	}
	grpcServer, err := grpc_adapter.NewGRPCAdapter(opts)
	if err != nil {
		logger.Error("Failed to create gRPC adapter:", "error", err)
		return
	}
	grpcServer.RegisterService(&userv1.UserService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx)

	grpcServer.Run(ctx)
}
