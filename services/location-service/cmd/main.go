package cmd

import (
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	grpcHandler "github.com/nepeta70/ride-hailing/services/location-service/internal/adapters/grpc"
	grcAdapters "github.com/nepeta70/ride-hailing/services/location-service/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//var shutDownDuration time.Duration = 5 * time.Second

func main() {
	// cfg, err := config.Load("config.json")
	// if err != nil {
	// 	log.Printf("Warning: could not load config.json (%v), using default port %d", err, cfg.Port)
	// }

	// logger := cfg.Logging.ConfigureLogger()

	// logger.Debug("Running Application mode: ", "applicationTType", cfg.ApplicationType, "eventStorageType", cfg.EventStorageType, "port", cfg.Port, "storageType", cfg.EventStorageType)

	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer stop()
	grpcServer := grpc.NewServer()

	locationRepository := grcAdapters.NewInMemoryLocationRepo()
	locationService := service.NewLocationService(locationRepository)
	handler := grpcHandler.NewLocationHandler(locationService)

	locationv1.RegisterLocationServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// <-ctx.Done()
	// logger.Debug("shutting down...")

	grpcServer.GracefulStop()

	// logger.Debug("shutdown complete")
}
