package grpc

import (
	"time"

	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Clients struct {
	Ride     ridev1.RideServiceClient
	User     userv1.UserServiceClient
	Driver   driverv1.DriverServiceClient
	Location locationv1.LocationServiceClient
	conns    []*grpc.ClientConn
}

func NewClients(cfg *config.Config, telemetry ports.TelemetryProvider) (*Clients, error) {
	rideConn, err := dial(cfg.Services.Ride.Address)
	if err != nil {
		telemetry.Metrics().DependencyFailure("gateway", "ride_client", err.Error())
		return nil, err
	}

	userConn, err := dial(cfg.Services.User.Address)
	if err != nil {
		telemetry.Metrics().DependencyFailure("gateway", "user_client", err.Error())
		_ = rideConn.Close()
		return nil, err
	}

	driverConn, err := dial(cfg.Services.Driver.Address)
	if err != nil {
		telemetry.Metrics().DependencyFailure("gateway", "driver_client", err.Error())
		_ = rideConn.Close()
		_ = userConn.Close()
		return nil, err
	}

	locationConn, err := dial(cfg.Services.Location.Address)
	if err != nil {
		telemetry.Metrics().DependencyFailure("gateway", "location_client", err.Error())
		_ = rideConn.Close()
		_ = userConn.Close()
		_ = driverConn.Close()
		return nil, err
	}

	return &Clients{
		Ride:     ridev1.NewRideServiceClient(rideConn),
		User:     userv1.NewUserServiceClient(userConn),
		Driver:   driverv1.NewDriverServiceClient(driverConn),
		Location: locationv1.NewLocationServiceClient(locationConn),
		conns:    []*grpc.ClientConn{rideConn, userConn, driverConn, locationConn},
	}, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}

func dial(address string) (*grpc.ClientConn, error) {
	retryPolicy := `{
		"methodConfig": [{
			"retryPolicy": {
				"maxAttempts": 5,
				"initialBackoff": "0.1s",
				"maxBackoff": "1s",
				"backoffMultiplier": 2,
				"retryableStatusCodes": ["UNAVAILABLE"]
			}
		}]
	}`

	return grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
	)
}
