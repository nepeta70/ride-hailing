package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	// Import your generated code
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
)

type LocationClient struct {
	client            locationv1.LocationServiceClient
	conn              *grpc.ClientConn
	telemetryProvider pkgPorts.TelemetryProvider
}

func NewLocationClient(address string, telemetryProvider pkgPorts.TelemetryProvider) (*LocationClient, error) {
	// Use insecure for internal cluster traffic, or set up mTLS
	retryPolicy := `{
        "methodConfig": [{
            "name": [{"service": "location.v1.LocationService"}],
            "retryPolicy": {
                "maxAttempts": 5,
                "initialBackoff": "0.1s",
                "maxBackoff": "1s",
                "backoffMultiplier": 2,
                "retryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
		// This ensures the client doesn't keep a "stale" dead connection
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		telemetryProvider.Metrics().DependencyFailure("LocationClient", "NewClient", err.Error())
		telemetryProvider.Logger().Error("Failed to create gRPC client", "error", err)
		return nil, err
	}
	return &LocationClient{
		client:            locationv1.NewLocationServiceClient(conn),
		conn:              conn,
		telemetryProvider: telemetryProvider,
	}, nil
}

func (lc *LocationClient) GetCandidates(ctx context.Context, coords *domain.Coordinates) ([]*locationv1.SearchNearbyDriversResponse_Driver, error) {
	req := &locationv1.SearchNearbyDriversRequest{
		Latitude:  coords.Latitude,
		Longitude: coords.Longitude,
	}

	// The call to Location Service
	resp, err := lc.client.SearchNearbyDrivers(ctx, req, grpc.WaitForReady(true))
	if err != nil {
		lc.telemetryProvider.Metrics().DependencyFailure("LocationClient", "SearchNearbyDrivers", err.Error())
		lc.telemetryProvider.Logger().ErrorContext(ctx, "Failed to call SearchNearbyDrivers", "error", err)
		return nil, err
	}

	return resp.Drivers, nil
}

func (lc *LocationClient) HealthCheck(ctx context.Context) error {
	return nil
}

func (lc *LocationClient) Close() {
	lc.conn.Close()
}

var _ ports.GetCandidates = (*LocationClient)(nil)
