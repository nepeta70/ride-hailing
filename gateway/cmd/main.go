package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	gatewayHTTP "github.com/nepeta70/ride-hailing/gateway/internal/adapters/http"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: could not load config.json (%v), using defaults on port %d", err, cfg.Server.Port)
	}

	tel, err := telemetry.NewTelemetryProvider(ctx, &cfg.BaseConfig)
	if err != nil {
		log.Printf("ERROR: Failed to create telemetry provider: %v", err)
		return
	}
	defer tel.Shutdown(ctx)

	clients, err := grpcClients.NewClients(cfg, tel)
	if err != nil {
		tel.Logger().Error("Failed to create gRPC clients", "error", err)
		return
	}
	defer clients.Close()

	server := gatewayHTTP.NewServer(&gatewayHTTP.Options{
		Config:    cfg,
		Clients:   clients,
		Telemetry: tel,
	})

	server.Run(ctx)
}
