package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	// Import your generated code
	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
)

type LocationClient struct {
	config            *config.Config
	client            locationv1.LocationServiceClient
	conn              *grpc.ClientConn
	telemetryProvider pkgPorts.TelemetryProvider
}

func NewLocationClient(config *config.Config, telemetryProvider pkgPorts.TelemetryProvider) (*LocationClient, error) {
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

	conn, err := grpc.NewClient(config.LocationService.LocationServiceAddress,
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
		config:            config,
	}, nil
}

func (lc *LocationClient) GetCandidates(ctx context.Context, coords *domain.Coordinates, headers map[string]string) ([]*locationv1.SearchNearbyDriversResponse_Driver, error) {
	ctx, span := lc.telemetryProvider.Tracer().Start(ctx, "LocationClient.GetCandidates",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("coordinates", coords.String()),
		),
	)
	defer span.End()

	ctx, err := lc.addMetadata(ctx, headers)
	if err != nil {
		span.RecordError(err)
		lc.telemetryProvider.Metrics().DependencyFailure("LocationClient", "addMetadata", err.Error())
		lc.telemetryProvider.Logger().ErrorContext(ctx, "Failed to add metadata to gRPC call", "error", err)
		return nil, err
	}

	req := &locationv1.SearchNearbyDriversRequest{
		Latitude:  coords.Latitude,
		Longitude: coords.Longitude,
	}

	// The call to Location Service
	resp, err := lc.client.SearchNearbyDrivers(ctx, req, grpc.WaitForReady(true))
	if err != nil {
		span.RecordError(err)
		lc.telemetryProvider.Metrics().DependencyFailure("LocationClient", "SearchNearbyDrivers", err.Error())
		lc.telemetryProvider.Logger().ErrorContext(ctx, "Failed to call SearchNearbyDrivers", "error", err)
		return nil, err
	}

	return resp.Drivers, nil
}

func (lc *LocationClient) UpdateDriverStatus(ctx context.Context, driverID uuid.UUID, status contracts.DriverStatus, headers map[string]string) error {
	ctx, span := lc.telemetryProvider.Tracer().Start(ctx, "LocationClient.GetCandidates",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("driver_id", driverID.String()),
			attribute.String("status", status.String()),
		),
	)
	defer span.End()
	ctx, err := lc.addMetadata(ctx, headers)
	if err != nil {
		span.RecordError(err)
		lc.telemetryProvider.Metrics().DependencyFailure("LocationClient", "addMetadata", err.Error())
		lc.telemetryProvider.Logger().ErrorContext(ctx, "Failed to add metadata to gRPC call", "error", err)
		return err
	}

	req := &locationv1.DriverStatus{
		DriverId: driverID.String(),
		Status:   status.String(),
	}
	_, err = lc.client.UpdateDriverStatus(ctx, req, grpc.WaitForReady(true))
	if err != nil {
		span.RecordError(err)
		lc.telemetryProvider.Metrics().DependencyFailure("LocationClient", "UpdateDriverStatus", err.Error())
		lc.telemetryProvider.Logger().ErrorContext(ctx, "Failed to call UpdateDriverStatus", "error", err)
		return err
	}
	return nil
}

func (lc *LocationClient) addMetadata(ctx context.Context, headers map[string]string) (context.Context, error) {
	md := metadata.New(headers)
	md.Set("timestamp", time.Now().UTC().Format(time.RFC3339))
	md.Set("api-key", lc.config.LocationService.APIKey)
	lc.telemetryProvider.Logger().DebugContext(ctx, "Adding metadata to outgoing gRPC call", "metadata", md)
	return metadata.NewOutgoingContext(ctx, md), nil
}

func (lc *LocationClient) HealthCheck(ctx context.Context) error {
	return nil
}

func (lc *LocationClient) Close() {
	lc.conn.Close()
}

var _ ports.GetCandidates = (*LocationClient)(nil)
