package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	caterrors "github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/service"
)

// LocationHandler implements the LocationService gRPC interface.
type LocationHandler struct {
	locationv1.UnimplementedLocationServiceServer
	app *app.Application
}

func NewLocationHandler(app *app.Application) *LocationHandler {
	return &LocationHandler{app: app}
}

func (h *LocationHandler) UpdateDriverLocation(ctx context.Context, req *locationv1.DriverLocation) (*emptypb.Empty, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	h.app.Logger.Debug("User updates location", "user-type", info.User.Role, "user-id", info.User.ID, "latitude", req.GetLatitude(), "longitude", req.GetLongitude())
	err = h.app.LocationService.Update(ctx, &service.UpdateRequest{
		UserID:   info.User.ID,
		UserType: enums.UserType(info.User.Role),
		Coordinates: domain.Coordinates{
			Latitude:  req.GetLatitude(),
			Longitude: req.GetLongitude(),
		},
		Accuracy:   req.GetAccuracy(),
		Heading:    req.GetHeading(),
		Speed:      req.GetSpeed(),
		CapturedAt: time.Unix(info.Trace.Timestamp, 0),
		Status:     contracts.DriverStatus(req.GetStatus()),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetDriverLocation handles the read-side (Query)
func (h *LocationHandler) GetDriverLocation(ctx context.Context, req *locationv1.UserID) (*locationv1.DriverLocation, error) {
	// 2. Execute Query
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}
	result, err := h.app.LocationService.Get(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	// 3. Map Result to Proto Response
	return &locationv1.DriverLocation{
		Latitude:  result.Coordinates.Latitude,
		Longitude: result.Coordinates.Longitude,
		Accuracy:  result.Accuracy,
		Heading:   result.Heading,
		Speed:     result.Speed,
	}, nil
}

// DeleteDriverLocation handles the deletion of a driver's location
func (h *LocationHandler) DeleteDriverLocation(ctx context.Context, req *locationv1.UserID) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}
	err = h.app.LocationService.RemoveUserLocation(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

// SearchNearbyDrivers handles searching for nearby drivers
func (h *LocationHandler) SearchNearbyDrivers(ctx context.Context, req *locationv1.SearchNearbyDriversRequest) (*locationv1.SearchNearbyDriversResponse, error) {
	results, err := h.app.LocationService.SearchNearby(ctx, &service.SearchNearbyRequest{
		Coordinates: domain.Coordinates{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		RadiusKm: req.RadiusKm,
	})
	if err != nil {
		return nil, mapError(err)
	}

	driverLocations := make([]*locationv1.DriverLocation, 0, len(results))
	for _, loc := range results {
		driverLocations = append(driverLocations, &locationv1.DriverLocation{
			Latitude:  loc.Coordinates.Latitude,
			Longitude: loc.Coordinates.Longitude,
			Accuracy:  loc.Accuracy,
			Heading:   loc.Heading,
			Speed:     loc.Speed,
		})
	}

	return &locationv1.SearchNearbyDriversResponse{
		DriverLocations: driverLocations,
	}, nil
}

func (h *LocationHandler) getInfoFromMetadata(ctx context.Context) (*ctxmgr.RequestInfo, error) {
	info, ok := h.app.ContextManager.Extract(ctx)

	if !ok {
		e := "no metadata found in context"
		h.app.Logger.Error(e)
		return nil, caterrors.NewPermanentError(e)
	}
	return info, nil
}

// mapError translates your internal CategorizedErrors to gRPC status codes
func mapError(err error) error {
	if caterrors.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, caterrors.ErrContextError) {
		return status.Error(codes.Canceled, "request cancelled or deadline exceeded")
	}
	if caterrors.IsBusiness(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if caterrors.IsTransient(err) {
		return status.Error(codes.Unavailable, "temporary service failure, retry allowed")
	}
	return status.Error(codes.Internal, "an internal error occurred")
}
