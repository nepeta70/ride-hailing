package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/grpc_adapter"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"
	grpcAdapters "github.com/nepeta70/ride-hailing/services/matching/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/services/matching/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/service"
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

	logger := tel.GetLogger()

	retrierFactory := retry.NewRetrierFactory(logger, tel.GetMetrics())

	contextManager := ctxmgr.NewContextManager()
	topicProvider := service.NewTopicProvider()

	publisher, err := pubsub.NewEventPublisher(&pubsub.KafkaPublisherOptions{
		Config:         cfg.Kafka,
		TopicProvider:  topicProvider,
		Logger:         logger,
		Metrics:        tel.GetMetrics(),
		RetrierFactory: retrierFactory,
		ContextManager: contextManager,
	})
	if err != nil {
		logger.Error("Failed to create event publisher:", "error", err)
		return
	}
	defer publisher.Close()

	subscriber, err := pubsub.NewKafkaSubscriber(&pubsub.KafkaSubscriberOptions{
		Config:         cfg.Kafka,
		GroupID:        cfg.ServiceName,
		Logger:         logger,
		RetrierFactory: retrierFactory,
		Metrics:        tel.GetMetrics(),
	})
	if err != nil {
		logger.Error("Failed to create event subscriber:", "error", err)
		return
	}
	defer subscriber.Close()

	locationClient := grpcAdapters.NewLocationClient(cfg.LocationService.LocationServiceAddress)
	defer locationClient.Close()

	matchingService, err := service.NewMatchingService(&service.MatchingServiceOpts{
		Config:         cfg,
		Client:         locationClient,
		Publisher:      publisher,
		Logger:         logger,
		Metrics:        tel.GetMetrics(),
		ContextManager: contextManager,
	})
	if err != nil {
		logger.Error("Failed to create matching service:", "error", err)
		return
	}

	application, err := app.NewApplication(&app.AppOptions{
		Logger:         logger,
		Metrics:        tel.GetMetrics(),
		Service:        matchingService,
		Subscriber:     subscriber,
		EventPublisher: publisher,
		ContextManager: contextManager,
	})
	if err != nil {
		logger.Error("Failed to create application:", "error", err)
		return
	}
	err = application.Start(ctx)
	if err != nil {
		logger.Error("Failed to start application:", "error", err)
		return
	}
	handler := grpcAdapters.NewMatchingHandler(application)

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
	grpcServer.RegisterService(&matchingv1.MatchingService_ServiceDesc, handler)

	grpcServer.MonitorHealth(ctx, publisher)

	grpcServer.Run(ctx)
}
