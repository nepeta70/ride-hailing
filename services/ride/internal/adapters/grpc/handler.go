package grpc

import (
	"context"

	"github.com/google/uuid"
	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/validator"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RideHandler implements the RideService gRPC interface.
type RideHandler struct {
	ridev1.UnimplementedRideServiceServer
	application   *app.Application
	storageBundle ports.StorageBundle
	grainSystem   *app.GrainSystem
	telemetry     pkgPorts.TelemetryProvider
}

func NewRideHandler(application *app.Application, storageBundle ports.StorageBundle, grainSystem *app.GrainSystem, telemetry pkgPorts.TelemetryProvider) *RideHandler {
	return &RideHandler{application: application, storageBundle: storageBundle, grainSystem: grainSystem, telemetry: telemetry}
}

func (h *RideHandler) EstimateFare(ctx context.Context, req *ridev1.FareEstimateRequest) (*ridev1.FareEstimateResponse, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	query := &commands.EstimateFareCommand{
		RequestID: info.Trace.RequestID,
		PickupLocation: &core.Coordinates{
			Latitude:  req.GetPickupLocation().GetLatitude(),
			Longitude: req.GetPickupLocation().GetLongitude(),
		},
		DropoffLocation: &core.Coordinates{
			Latitude:  req.GetDropoffLocation().GetLatitude(),
			Longitude: req.GetDropoffLocation().GetLongitude(),
		},
		CountryCode: info.Location.CountryCode,
		RiderID:     info.Sender.ID,
	}
	fare, err := h.application.Commands.EstimateFare.Handle(ctx, *query)
	if err != nil {
		return nil, err
	}
	n := len(fare.Fares)

	fareEstimates := make([]*ridev1.FareEstimate, 0, n)
	for s, f := range fare.Fares {
		fare := &ridev1.FareEstimate{
			ServiceType:   s,
			EstimatedFare: f,
		}
		fareEstimates = append(fareEstimates, fare)
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
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	val := validator.New()
	val.StringField("fare_id", req.GetFareId()).Required().IsUUID()
	if val.HasErrors() {
		return nil, val.Errors()
	}

	cmd := commands.RequestRide{
		FareID:      uuid.MustParse(req.GetFareId()),
		ServiceType: req.GetServiceType(),
		RiderID:     info.Sender.ID,
		RequestID:   info.Trace.RequestID,
	}
	ride, err := h.application.Commands.RequestRide.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &ridev1.RideResponse{
		RideId: ride.RideID.String(),
	}, nil
}

func (h *RideHandler) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*emptypb.Empty, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	val := validator.New()
	val.StringField("ride_id", req.GetRideId()).Required().IsUUID()
	if val.HasErrors() {
		return nil, val.Errors()
	}

	cmd := &grains.CancelRideCommand{
		RequestID: info.Trace.RequestID,
		RiderID:   info.Sender.ID,
		RideID:    uuid.MustParse(req.GetRideId()),
	}

	identity := grain.NewGrainIdentity(domain.RideGrainKind, cmd.RideID)

	_, err = h.grainSystem.Silo().Ask(ctx, identity, cmd)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *RideHandler) AcceptOrRejectRide(ctx context.Context, req *ridev1.AcceptOrRejectRideRequest) (*emptypb.Empty, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	val := validator.New()
	val.StringField("ride_id", req.GetRideId()).Required().IsUUID()
	if val.HasErrors() {
		return nil, val.Errors()
	}

	rideID := uuid.MustParse(req.GetRideId())

	var cmd pkgPorts.MessageInterface
	if req.GetAccept() {
		cmd = &grains.AcceptRideCommand{
			RequestID: info.Trace.RequestID,
			DriverID:  info.Sender.ID,
			RideID:    rideID,
		}
	} else {
		cmd = &grains.RejectRideCommand{
			RequestID: info.Trace.RequestID,
			DriverID:  info.Sender.ID,
			RideID:    rideID,
		}
	}

	identity := grain.NewGrainIdentity(domain.RideGrainKind, rideID)

	_, err = h.grainSystem.Silo().Ask(ctx, identity, cmd)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *RideHandler) StartRide(ctx context.Context, req *ridev1.StartRideRequest) (*emptypb.Empty, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	val := validator.New()
	val.StringField("ride-id", req.GetRideId()).Required().IsUUID()
	if val.HasErrors() {
		return nil, val.Errors()
	}

	cmd := &grains.StartRideCommand{
		RideID:    uuid.MustParse(req.GetRideId()),
		DriverID:  info.Sender.ID,
		RequestID: info.Trace.RequestID,
	}
	identity := grain.NewGrainIdentity(domain.RideGrainKind, cmd.RideID)

	_, err = h.grainSystem.Silo().Ask(ctx, identity, cmd)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *RideHandler) CompleteRide(ctx context.Context, req *ridev1.CompleteRideRequest) (*emptypb.Empty, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	val := validator.New()
	val.StringField("ride-id", req.GetRideId()).Required().IsUUID()
	if val.HasErrors() {
		return nil, val.Errors()
	}

	cmd := &grains.CompleteRideCommand{
		RideID:    uuid.MustParse(req.GetRideId()),
		DriverID:  info.Sender.ID,
		RequestID: info.Trace.RequestID,
	}
	identity := grain.NewGrainIdentity(domain.RideGrainKind, cmd.RideID)

	_, err = h.grainSystem.Silo().Ask(ctx, identity, cmd)
	if err != nil {
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

func (h *RideHandler) getInfoFromMetadata(ctx context.Context) (*ctxmgr.RequestInfo, error) {
	info, ok := h.application.ContextManager.Extract(ctx)

	if !ok {
		e := "no metadata found in context"
		h.telemetry.Logger().ErrorContext(ctx, e)
		return nil, errors.NewPermanentError(e)
	}
	return info, nil
}
