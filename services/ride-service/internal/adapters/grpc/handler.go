package grpc

import (
	"context"

	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RideHandler implements the RideService gRPC interface.
type RideHandler struct {
	ridev1.UnimplementedRideServiceServer
	application *app.Application
	fareService *service.FareService
	rideService *service.RideService
}

func NewRideHandler(application *app.Application, fareService *service.FareService, rideService *service.RideService) *RideHandler {
	return &RideHandler{application: application, fareService: fareService, rideService: rideService}
}

func (h *RideHandler) RequestFareEstimate(ctx context.Context, req *ridev1.FareEstimateRequest) (*ridev1.FareEstimateResponse, error) {
	query := &queries.GetFare{
		PickupLocation:  req.PickupLocation,
		DropoffLocation: req.DropoffLocation,
		CountryCode:     req.CountryCode,
	}
	fare, err := h.application.Queries.GetFare.Handle(ctx, *query)
	if err != nil {
		return nil, err
	}
	return &ridev1.FareEstimateResponse{
		FareId:                   fare.ID.String(),
		EstimatedDistanceKm:      int32(fare.EstimatedDistanceKm),
		EstimatedDurationMinutes: int32(fare.EstimatedDuration.Minutes()),
		EstimatedFare:            fare.Fare,
		Currency:                 fare.Currency,
	}, nil
}

func (h *RideHandler) RequestRide(ctx context.Context, req *ridev1.RideRequest) (*ridev1.RideResponse, error) {
	// TODO: Call the ride service logic here
	return &ridev1.RideResponse{}, nil
}

func (h *RideHandler) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*emptypb.Empty, error) {
	// TODO: Call the ride service logic here
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) AcceptOrRejectRide(ctx context.Context, req *ridev1.AcceptOrRejectRideRequest) (*emptypb.Empty, error) {
	// TODO: Call the ride service logic here
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) StartRide(ctx context.Context, req *ridev1.StartRideRequest) (*emptypb.Empty, error) {
	// TODO: Call the ride service logic here
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) CompleteRide(ctx context.Context, req *ridev1.CompleteRideRequest) (*emptypb.Empty, error) {
	// TODO: Call the ride service logic here
	return &emptypb.Empty{}, nil
}
