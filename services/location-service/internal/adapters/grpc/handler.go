package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
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

func (h *LocationHandler) UpdateLocation(ctx context.Context, req *locationv1.UpdateLocationRequest) (*emptypb.Empty, error) {
	err := h.service.Update(ctx, &service.UpdateRequest{
		EntityID:   req.EntityId,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
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
func (h *LocationHandler) GetLocation(ctx context.Context, req *locationv1.LocateRequest) (*locationv1.Location, error) {
	// 2. Execute Query
	result, err := h.service.Get(ctx, req.EntityId)
	if err != nil {
		return nil, mapError(err)
	}

	// 3. Map Result to Proto Response
	return &locationv1.Location{
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
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
