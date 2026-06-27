package grpc

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	pkgErrors "github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/service"
)

// LocationHandler implements the LocationService gRPC interface.
type LocationHandler struct {
	//locationv1.UnimplementedLocationServiceServer
	app       *app.Application
	telemetry ports.TelemetryProvider
}

func NewLocationHandler(app *app.Application, telemetry ports.TelemetryProvider) *LocationHandler {
	return &LocationHandler{app: app, telemetry: telemetry}
}

func (h *LocationHandler) UpdateDriverLocation(ctx context.Context, req *locationv1.DriverLocation) (*emptypb.Empty, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "UpdateDriverLocation",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("sender.id", info.Sender.ID.String()),
		attribute.String("sender.type", info.Sender.Type.String()),
		attribute.Float64("location.latitude", req.GetLatitude()),
		attribute.Float64("location.longitude", req.GetLongitude()),
	)

	h.telemetry.Logger().DebugContext(ctx, "User updates location.", "user-type", info.Sender.Type, "sender-id", info.Sender.ID, "latitude", req.GetLatitude(), "longitude", req.GetLongitude())

	updateRequest := &service.UpdateDriverLocationRequest{
		DriverID:   info.Sender.ID,
		SenderType: info.Sender.Type,
		Coordinates: common.Coordinates{
			Latitude:  req.GetLatitude(),
			Longitude: req.GetLongitude(),
		},
		Accuracy:   req.GetAccuracy(),
		Heading:    req.GetHeading(),
		Speed:      req.GetSpeed(),
		CapturedAt: info.Trace.Timestamp,
	}

	err = updateRequest.Validate()
	if err != nil {
		span.RecordError(err)
		return nil, mapError(err)
	}
	err = h.app.LocationService.Update(ctx, updateRequest)
	if err != nil {
		span.RecordError(err)
		return nil, mapError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetDriverLocation handles the read-side (Query)
func (h *LocationHandler) GetDriverLocation(ctx context.Context, req *locationv1.UserID) (*locationv1.DriverLocationWithStatus, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "UpdateDriverLocation",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		span.RecordError(err)
		return nil, mapError(pkgErrors.NewValidationErrorf("invalid user ID format"))
	}
	result, err := h.app.LocationService.Get(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, mapError(err)
	}

	// 3. Map Result to Proto Response
	return &locationv1.DriverLocationWithStatus{
		Location: &locationv1.DriverLocation{
			Latitude:  result.Coordinates.Latitude,
			Longitude: result.Coordinates.Longitude,
			Accuracy:  result.Accuracy,
			Heading:   result.Heading,
			Speed:     result.Speed,
		},
		Status:          result.Status.String(),
		StatusUpdatedAt: timestamppb.New(result.StatusUpdatedAt),
	}, nil
}

// DeleteDriverLocation handles the deletion of a driver's location
func (h *LocationHandler) DeleteDriverLocation(ctx context.Context, req *locationv1.UserID) (*emptypb.Empty, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "DeleteDriverLocation",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("driver.id", req.UserId),
		),
	)
	defer span.End()

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, mapError(pkgErrors.NewValidationErrorf("invalid user ID format"))
	}
	err = h.app.LocationService.RemoveUserLocation(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

// SearchNearbyDrivers handles searching for nearby drivers
func (h *LocationHandler) SearchNearbyDrivers(ctx context.Context, req *locationv1.SearchNearbyDriversRequest) (*locationv1.SearchNearbyDriversResponse, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "SearchNearbyDrivers",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.Float64("latitude", req.GetLatitude()),
			attribute.Float64("longitude", req.GetLongitude()),
		),
	)
	defer span.End()

	coords := &common.Coordinates{
		Latitude:  req.GetLatitude(),
		Longitude: req.GetLongitude(),
	}
	err := coords.Validate()
	if err != nil {
		span.RecordError(err)
		return nil, mapError(pkgErrors.NewValidationErrorf("invalid coordinates: %w", err))
	}

	results, err := h.app.LocationService.SearchNearby(ctx, coords, req.GetRadiusKm())
	if err != nil {
		span.RecordError(err)
		return nil, mapError(err)
	}

	driverLocations := make([]*locationv1.SearchNearbyDriversResponse_Driver, 0, len(results.Drivers))
	for _, loc := range results.Drivers {
		driverLocations = append(driverLocations, &locationv1.SearchNearbyDriversResponse_Driver{
			UserId:         loc.UserID.String(),
			Latitude:       loc.Coordinates.Latitude,
			Longitude:      loc.Coordinates.Longitude,
			DistanceKm:     loc.DistanceKm,
			AvailableSince: timestamppb.New(loc.StatusUpdatedAt),
		})
	}

	span.AddEvent("Found nearby drivers", trace.WithAttributes(
		attribute.Int("driver.count", len(driverLocations)),
	))
	return &locationv1.SearchNearbyDriversResponse{
		RadiusKm: results.RadiusKm,
		Drivers:  driverLocations,
	}, nil
}

func (h *LocationHandler) UpdateDriverStatus(ctx context.Context, req *locationv1.DriverStatus) (*emptypb.Empty, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "UpdateDriverStatus",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("driver.id", req.DriverId),
			attribute.String("status", req.Status),
		),
	)
	defer span.End()
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	h.telemetry.Logger().DebugContext(ctx, "Updating driver status.", "driver_id", req.DriverId, "status", req.Status, "sender_id", info.Sender.ID, "sender_type", info.Sender.Type)
	driverID, err := uuid.Parse(req.GetDriverId())
	if err != nil {
		span.RecordError(err)
		h.telemetry.Logger().ErrorContext(ctx, "Invalid driver ID format", "driver_id", req.DriverId)
		return nil, mapError(pkgErrors.NewValidationErrorf("invalid driver ID format"))
	}
	newStatus := contracts.DriverStatus(req.GetStatus())
	if !newStatus.IsValid() {
		span.RecordError(errors.New("invalid driver status value"))
		h.telemetry.Logger().ErrorContext(ctx, "Invalid driver status value", "driver_id", req.DriverId, "status", req.GetStatus())
		return nil, mapError(pkgErrors.NewValidationErrorf("invalid driver status value"))
	}

	updateRequest := &service.UpdateDriverStatusRequest{
		DriverID:        driverID,
		Status:          newStatus,
		StatusUpdatedAt: time.Now().UTC(),
	}
	err = updateRequest.Validate()
	if err != nil {
		span.RecordError(err)
		h.telemetry.Logger().ErrorContext(ctx, "Invalid driver status", "driver_id", req.DriverId, "status", req.GetStatus(), "error", err)
		return nil, mapError(err)
	}
	err = h.app.LocationService.UpdateDriverStatus(ctx, updateRequest)
	if err != nil {
		h.telemetry.Logger().ErrorContext(ctx, "Failed to update driver status", "driver_id", req.DriverId, "status", req.GetStatus(), "error", err)
		span.RecordError(err)
		return nil, mapError(err)
	}

	span.SetStatus(otelCodes.Ok, "Driver status updated successfully")
	return &emptypb.Empty{}, nil
}

func (h *LocationHandler) getInfoFromMetadata(ctx context.Context) (*ctxmgr.RequestInfo, error) {
	info, ok := h.app.ContextManager.Extract(ctx)

	if !ok {
		e := "no metadata found in context"
		h.telemetry.Logger().ErrorContext(ctx, e)
		return nil, pkgErrors.NewPermanentError(e)
	}
	return info, nil
}

// mapError translates your internal CategorizedErrors to gRPC status codes
func mapError(err error) error {
	if pkgErrors.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, pkgErrors.ErrContextError) {
		return status.Error(codes.Canceled, "request cancelled or deadline exceeded")
	}
	if pkgErrors.IsBusiness(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if pkgErrors.IsTransient(err) {
		return status.Error(codes.Unavailable, "temporary service failure, retry allowed")
	}
	return status.Error(codes.Internal, "an internal error occurred")
}

var _ locationv1.LocationServiceServer = (*LocationHandler)(nil)
