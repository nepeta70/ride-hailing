package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/handlers"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/middleware"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Server struct {
	engine *gin.Engine
	server *http.Server
	logger ports.Logger
}

type Options struct {
	Config    *config.Config
	Clients   *grpcClients.Clients
	Telemetry ports.TelemetryProvider
}

func NewServer(opts *Options) *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(otelgin.Middleware(opts.Config.ServiceName,
		otelgin.WithTracerProvider(opts.Telemetry.TracerProvider()),
		otelgin.WithMeterProvider(opts.Telemetry.MeterProvider()),
		otelgin.WithPropagators(opts.Telemetry.Propagator()),
	))
	engine.Use(middleware.Recovery(opts.Telemetry))
	engine.Use(middleware.RequestContextMiddleware(&middleware.HTTPMiddlewareOptions{
		Telemetry: opts.Telemetry,
		Config:    opts.Config,
	}))
	engine.Use(gin.Logger())

	rideHandler := handlers.NewRideHandler(opts.Clients, opts.Config, opts.Telemetry)
	userHandler := handlers.NewUserHandler(opts.Clients, opts.Config, opts.Telemetry)
	driverHandler := handlers.NewDriverHandler(opts.Clients, opts.Config, opts.Telemetry)
	locationHandler := handlers.NewLocationHandler(opts.Clients, opts.Config, opts.Telemetry)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": opts.Config.ServiceName})
	})

	v1 := engine.Group("/api/v1")
	{
		rides := v1.Group("/rides")
		{
			rides.POST("/estimate", rideHandler.EstimateFare)
			rides.POST("", rideHandler.RequestRide)
			rides.POST("/:id/cancel", rideHandler.CancelRide)
			rides.POST("/:id/accept-reject", rideHandler.AcceptOrRejectRide)
			rides.POST("/:id/start", rideHandler.StartRide)
			rides.POST("/:id/complete", rideHandler.CompleteRide)
		}

		fareRates := v1.Group("/fare-rates")
		{
			fareRates.GET("", rideHandler.GetFareRates)
			fareRates.POST("", rideHandler.CreateFareRate)
			fareRates.PUT("/:id", rideHandler.UpdateFareRate)
		}

		users := v1.Group("/users")
		{
			users.POST("", userHandler.RegisterUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
		}

		drivers := v1.Group("/drivers")
		{
			drivers.POST("", driverHandler.CreateDriver)
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.PUT("/:id", driverHandler.UpdateDriver)
		}

		location := v1.Group("/location")
		{
			location.PUT("/drivers/me", locationHandler.UpdateDriverLocation)
			location.GET("/drivers/:id", locationHandler.GetDriverLocation)
			location.DELETE("/drivers/:id", locationHandler.DeleteDriverLocation)
			location.GET("/drivers/nearby", locationHandler.SearchNearbyDrivers)
			location.PUT("/drivers/:id/status", locationHandler.UpdateDriverStatus)
		}
	}

	addr := opts.Config.Server.Host + ":" + strconv.Itoa(opts.Config.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  opts.Config.Server.ReadTimeout,
		WriteTimeout: opts.Config.Server.WriteTimeout,
		IdleTimeout:  opts.Config.Server.IdleTimeout,
	}

	return &Server{
		engine: engine,
		server: httpServer,
		logger: opts.Telemetry.Logger(),
	}
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		s.logger.InfoContext(ctx, "HTTP gateway starting", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "HTTP gateway error", "error", err)
		}
	}()

	<-ctx.Done()
	s.logger.InfoContext(ctx, "shutting down HTTP gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(ctx, "HTTP gateway shutdown error", "error", err)
	}
}
