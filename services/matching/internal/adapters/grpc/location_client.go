package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import your generated code
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
)

type LocationClient struct {
	client locationv1.LocationServiceClient
	conn   *grpc.ClientConn
}

func NewLocationClient(address string) *LocationClient {
	// Use insecure for internal cluster traffic, or set up mTLS
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to location service: %v", err)
	}
	return &LocationClient{
		client: locationv1.NewLocationServiceClient(conn),
		conn:   conn,
	}
}

func (lc *LocationClient) GetCandidates(ctx context.Context, coords *domain.Coordinates) ([]*locationv1.SearchNearbyDriversResponse_Driver, error) {
	req := &locationv1.SearchNearbyDriversRequest{
		Latitude:  coords.Latitude,
		Longitude: coords.Longitude,
	}

	// The call to Location Service
	resp, err := lc.client.SearchNearbyDrivers(ctx, req)
	if err != nil {
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
