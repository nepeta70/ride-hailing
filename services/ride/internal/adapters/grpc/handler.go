package grpc

import (
	"context"

	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	ridePorts "github.com/nepeta70/ride-hailing/services/ride/internal/ports"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RideHandler implements the RideService gRPC interface.
type RideHandler struct {
	ridev1.UnimplementedRideServiceServer
	application   *app.Application
	storageBundle ridePorts.StorageBundle
}

func NewRideHandler(application *app.Application, storageBundle ridePorts.StorageBundle) *RideHandler {
	return &RideHandler{application: application, storageBundle: storageBundle}
}

func (h *RideHandler) EstimateFare(ctx context.Context, req *ridev1.FareEstimateRequest) (*ridev1.FareEstimateResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		h.application.Logger.Info("Received metadata: %v", md)
	}

	info, ok := h.application.ContextManager.Extract(ctx)

	if !ok {
		h.application.Logger.Error("No country code found in context.")
		return nil, errors.NewPermanentError("no country code found in context")
	}

	query := &commands.EstimateFareCommand{
		PickupLocation:  req.PickupLocation,
		DropoffLocation: req.DropoffLocation,
		CountryCode:     info.Location.CountryCode,
	}
	fare, err := h.application.Commands.FareEstimate.Handle(ctx, *query)
	if err != nil {
		return nil, err
	}
	n := len(fare.Fares)
	fareEstimates := make([]*ridev1.FareEstimate, n)
	for i, f := range fare.Fares {
		fareEstimates[i] = &ridev1.FareEstimate{
			ServiceType:   f.ServiceType,
			EstimatedFare: f.Fare,
		}
	}

	return &ridev1.FareEstimateResponse{
		Id:                       fare.ID.String(),
		EstimatedDistanceKm:      int32(fare.EstimatedDistanceKm),
		EstimatedDurationMinutes: int32(fare.EstimatedDurationMinutes.Minutes()),
		Currency:                 fare.Currency,
		FareEstimate:             fareEstimates,
	}, nil

}

func (h *RideHandler) RequestRide(ctx context.Context, req *ridev1.RideRequest) (*ridev1.RideResponse, error) {
	cmd := commands.RequestRide{
		FareId: req.FareId,
	}
	ride, err := h.application.Commands.RequestRide.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &ridev1.RideResponse{
		RideId: ride.ID.String(),
	}, nil
}

func (h *RideHandler) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*emptypb.Empty, error) {
	cmd := commands.CancelRide{
		RideID: req.RideId,
	}
	if err := h.application.Commands.CancelRide.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) AcceptOrRejectRide(ctx context.Context, req *ridev1.AcceptOrRejectRideRequest) (*emptypb.Empty, error) {
	cmd := commands.AcceptOrRejectRide{
		RideID: req.RideId,
		Accept: req.Accept,
	}
	if err := h.application.Commands.AcceptOrRejectRide.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) StartRide(ctx context.Context, req *ridev1.StartRideRequest) (*emptypb.Empty, error) {
	cmd := commands.StartRide{
		RideID: req.RideId,
	}
	if err := h.application.Commands.StartRide.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) CompleteRide(ctx context.Context, req *ridev1.CompleteRideRequest) (*emptypb.Empty, error) {
	cmd := commands.CompleteRide{
		RideID: req.RideId,
	}
	if err := h.application.Commands.CompleteRide.Handle(ctx, cmd); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) CreateFareRate(ctx context.Context, req *ridev1.FareRate) (*ridev1.FareRate, error) {
	return nil, nil
}

func (h *RideHandler) GetFareRates(ctx context.Context, req *ridev1.GetFareRatesRequest) (*ridev1.GetFareRatesResponse, error) {
	query := &queries.GetFareRates{
		CountryCode: req.Country,
	}
	rates, err := h.application.Queries.FareRates.Handle(ctx, query)
	if err != nil {
		return nil, err
	}
	var resp []*ridev1.FareRate
	for _, rate := range rates {
		resp = append(resp, &ridev1.FareRate{
			Id:            rate.ID.String(),
			Country:       rate.CountryCode,
			BaseFare:      rate.BaseFare,
			CostPerKm:     rate.FarePerKm,
			CostPerMinute: rate.FarePerMinute,
			MinimumFare:   rate.MinimumFare,
			Currency:      rate.Currency,
		})
	}
	response := ridev1.GetFareRatesResponse{
		FareRates: resp,
	}
	return &response, nil
}

func (h *RideHandler) UpdateFareRate(ctx context.Context, req *ridev1.FareRate) (*ridev1.FareRate, error) {
	return nil, nil
}
