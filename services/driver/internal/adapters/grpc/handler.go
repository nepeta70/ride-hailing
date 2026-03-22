package grpc

import (
	"context"

	"github.com/google/uuid"
	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/service"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DriverHandler for gRPC requests (template)
type DriverHandler struct {
	driverv1.UnimplementedDriverServiceServer
	driverService  *service.DriverService
	telemetry      pkgPorts.TelemetryProvider
	contextManager *ctxmgr.ContextManager
}

func NewDriverHandler(driverService *service.DriverService, telemetry pkgPorts.TelemetryProvider, contextManager *ctxmgr.ContextManager) *DriverHandler {
	return &DriverHandler{
		driverService:  driverService,
		telemetry:      telemetry,
		contextManager: contextManager,
	}
}
func (h *DriverHandler) CreateDriver(ctx context.Context, req *driverv1.CreateDriverRequest) (*driverv1.Driver, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "CreateDriver",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	driver := &domain.Driver{
		UserID:        info.Sender.ID,
		LicenseNumber: req.GetLicenseNumber(),
		LicenseExpiry: req.GetLicenseNumberExpiryDate().AsTime(),
		Vehicle: &domain.Vehicle{
			Make:         req.GetVehicleInfo().GetMake(),
			Model:        req.GetVehicleInfo().GetModel(),
			Color:        req.GetVehicleInfo().GetColor(),
			LicensePlate: req.GetVehicleInfo().GetLicensePlate(),
			Seats:        req.GetVehicleInfo().GetSeats(),
			Category:     req.GetVehicleInfo().GetCategory(),
		},
	}
	driver, err = h.driverService.CreateDriver(ctx, driver)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetStatus(codes.Ok, "Driver created")
	return &driverv1.Driver{
		UserId:                  driver.UserID.String(),
		LicenseNumber:           driver.LicenseNumber,
		LicenseNumberExpiryDate: timestamppb.New(driver.LicenseExpiry),
		VehicleInfo: &driverv1.VehicleInfo{
			Make:              driver.Vehicle.Make,
			Model:             driver.Vehicle.Model,
			Color:             driver.Vehicle.Color,
			LicensePlate:      driver.Vehicle.LicensePlate,
			Seats:             driver.Vehicle.Seats,
			Category:          driver.Vehicle.Category,
			AcceptsPets:       driver.Vehicle.AcceptsPets,
			AcceptsWheelchair: driver.Vehicle.AcceptsWheelchair,
			AdditionalInfo:    driver.Vehicle.AdditionalInfo,
		},
	}, nil
}

func (h *DriverHandler) UpdateDriver(ctx context.Context, req *driverv1.Driver) (*driverv1.Driver, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "CreateDriver",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	driver := &domain.Driver{
		UserID:        info.Sender.ID,
		LicenseNumber: req.GetLicenseNumber(),
		LicenseExpiry: req.GetLicenseNumberExpiryDate().AsTime(),
		Vehicle: &domain.Vehicle{
			Make:         req.GetVehicleInfo().GetMake(),
			Model:        req.GetVehicleInfo().GetModel(),
			Color:        req.GetVehicleInfo().GetColor(),
			LicensePlate: req.GetVehicleInfo().GetLicensePlate(),
			Seats:        req.GetVehicleInfo().GetSeats(),
			Category:     req.GetVehicleInfo().GetCategory(),
		},
	}
	driver, err = h.driverService.UpdateDriver(ctx, driver)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetStatus(codes.Ok, "Driver updated")
	return &driverv1.Driver{
		UserId:                  driver.UserID.String(),
		LicenseNumber:           driver.LicenseNumber,
		LicenseNumberExpiryDate: timestamppb.New(driver.LicenseExpiry),
		VehicleInfo: &driverv1.VehicleInfo{
			Make:              driver.Vehicle.Make,
			Model:             driver.Vehicle.Model,
			Color:             driver.Vehicle.Color,
			LicensePlate:      driver.Vehicle.LicensePlate,
			Seats:             driver.Vehicle.Seats,
			Category:          driver.Vehicle.Category,
			AcceptsPets:       driver.Vehicle.AcceptsPets,
			AcceptsWheelchair: driver.Vehicle.AcceptsWheelchair,
			AdditionalInfo:    driver.Vehicle.AdditionalInfo,
		},
	}, nil
}

func (h *DriverHandler) GetDriver(ctx context.Context, req *driverv1.GetDriverRequest) (*driverv1.Driver, error) {
	info, err := h.getInfoFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	driverID := info.Sender.ID
	if info.Sender.Type == enums.SenderTypeAdmin {
		if driverID, err = uuid.Parse(req.GetUserId()); driverID == uuid.Nil {
			return nil, err
		}
	}

	driver, err := h.driverService.GetDriver(ctx, driverID)
	if err != nil {
		return nil, err
	}

	return &driverv1.Driver{
		UserId:                  driver.UserID.String(),
		LicenseNumber:           driver.LicenseNumber,
		LicenseNumberExpiryDate: timestamppb.New(driver.LicenseExpiry),
		VehicleInfo: &driverv1.VehicleInfo{
			Make:              driver.Vehicle.Make,
			Model:             driver.Vehicle.Model,
			Color:             driver.Vehicle.Color,
			LicensePlate:      driver.Vehicle.LicensePlate,
			Seats:             driver.Vehicle.Seats,
			Category:          driver.Vehicle.Category,
			AcceptsPets:       driver.Vehicle.AcceptsPets,
			AcceptsWheelchair: driver.Vehicle.AcceptsWheelchair,
			AdditionalInfo:    driver.Vehicle.AdditionalInfo,
		},
	}, nil
}

func (h *DriverHandler) getInfoFromMetadata(ctx context.Context) (*ctxmgr.RequestInfo, error) {
	info, ok := h.contextManager.Extract(ctx)

	if !ok {
		e := "no metadata found in context"
		h.telemetry.Logger().ErrorContext(ctx, e)
		return nil, errors.NewPermanentError(e)
	}
	return info, nil
}
