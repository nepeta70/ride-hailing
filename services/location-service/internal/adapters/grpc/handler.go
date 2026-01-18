package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/service"
)

// LocationHandler implements the LocationService gRPC interface.
type LocationHandler struct {
	locationv1.UnimplementedLocationServiceServer
	service *service.LocationService
}

func NewLocationHandler(service *service.LocationService) *LocationHandler {
	return &LocationHandler{service: service}
}

func (h *LocationHandler) UpdateLocation(ctx context.Context, req *locationv1.UpdateUserLocationRequest) (*emptypb.Empty, error) {
	err := h.service.Update(ctx, &service.UpdateRequest{
		UserID: req.UserId,
		Coordinates: domain.Coordinates{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		Accuracy:   req.Accuracy,
		Heading:    req.Heading,
		Speed:      req.Speed,
		CapturedAt: req.CapturedAt.AsTime(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetLocation handles the read-side (Query)
func (h *LocationHandler) GetLocation(ctx context.Context, req *locationv1.LocateUserLocationRequest) (*locationv1.UserLocation, error) {
	// 2. Execute Query
	result, err := h.service.Get(ctx, req.UserId)
	if err != nil {
		return nil, mapError(err)
	}

	// 3. Map Result to Proto Response
	return &locationv1.UserLocation{
		Latitude:   result.Coordinates.Latitude,
		Longitude:  result.Coordinates.Longitude,
		Accuracy:   result.Accuracy,
		Heading:    result.Heading,
		Speed:      result.Speed,
		CapturedAt: timestamppb.New(result.CapturedAt),
	}, nil
}

// DeleteUserLocation handles the deletion of a user's location
func (h *LocationHandler) DeleteUserLocation(ctx context.Context, req *locationv1.DeleteUserLocationRequest) (*emptypb.Empty, error) {
	err := h.service.RemoveUserLocation(ctx, req.UserId)
	if err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

// SearchNearbyDrivers handles searching for nearby drivers
func (h *LocationHandler) SearchNearbyDrivers(ctx context.Context, req *locationv1.SearchNearbyDriversRequest) (*locationv1.SearchNearbyDriversResponse, error) {
	results, err := h.service.SearchNearby(ctx, &service.SearchNearbyRequest{
		Coordinates: domain.Coordinates{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		RadiusKm: req.RadiusKm,
	})
	if err != nil {
		return nil, mapError(err)
	}

	driverLocations := make([]*locationv1.UserLocation, 0, len(results))
	for _, loc := range results {
		driverLocations = append(driverLocations, &locationv1.UserLocation{
			Latitude:   loc.Coordinates.Latitude,
			Longitude:  loc.Coordinates.Longitude,
			Accuracy:   loc.Accuracy,
			Heading:    loc.Heading,
			Speed:      loc.Speed,
			CapturedAt: timestamppb.New(loc.CapturedAt),
		})
	}

	return &locationv1.SearchNearbyDriversResponse{
		DriverLocations: driverLocations,
	}, nil
}

// mapError translates your internal CategorizedErrors to gRPC status codes
func mapError(err error) error {
	if errors.IsBusiness(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.IsTransient(err) {
		return status.Error(codes.Unavailable, "temporary service failure, retry allowed")
	}
	return status.Error(codes.Internal, "an internal error occurred")
}
