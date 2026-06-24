package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/middleware"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DriverHandler struct {
	client      driverv1.DriverServiceClient
	apiKey      string
	hmacSecret  string
	propagator  propagation.TextMapPropagator
}

func NewDriverHandler(clients *grpcClients.Clients, cfg *config.Config, telemetry ports.TelemetryProvider) *DriverHandler {
	return &DriverHandler{
		client:     clients.Driver,
		apiKey:     cfg.Services.Driver.APIKey,
		hmacSecret: cfg.HMACSecret,
		propagator: telemetry.Propagator(),
	}
}

type vehicleInfoPayload struct {
	Make              string `json:"make" binding:"required"`
	Model             string `json:"model" binding:"required"`
	Color             string `json:"color" binding:"required"`
	LicensePlate      string `json:"license_plate" binding:"required"`
	Seats             int32  `json:"seats" binding:"required"`
	Category          string `json:"category" binding:"required"`
	AcceptsPets       bool   `json:"accepts_pets"`
	AcceptsWheelchair bool   `json:"accepts_wheelchair"`
	AdditionalInfo    string `json:"additional_info"`
}

type createDriverPayload struct {
	UserID                   string             `json:"user_id" binding:"required"`
	LicenseNumber            string             `json:"license_number" binding:"required"`
	LicenseNumberExpiryDate  time.Time          `json:"license_number_expiry_date" binding:"required"`
	VehicleInfo              vehicleInfoPayload `json:"vehicle_info" binding:"required"`
}

func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var body createDriverPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.CreateDriver(ctx, &driverv1.CreateDriverRequest{
		UserId:                  body.UserID,
		LicenseNumber:           body.LicenseNumber,
		LicenseNumberExpiryDate: timestamppb.New(body.LicenseNumberExpiryDate),
		VehicleInfo:             vehiclePayloadToProto(&body.VehicleInfo),
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

type updateDriverPayload struct {
	LicenseNumber           string             `json:"license_number" binding:"required"`
	LicenseNumberExpiryDate time.Time          `json:"license_number_expiry_date" binding:"required"`
	VehicleInfo             vehicleInfoPayload `json:"vehicle_info" binding:"required"`
}

func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	var body updateDriverPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.UpdateDriver(ctx, &driverv1.Driver{
		UserId:                  c.Param("id"),
		LicenseNumber:           body.LicenseNumber,
		LicenseNumberExpiryDate: timestamppb.New(body.LicenseNumberExpiryDate),
		VehicleInfo:             vehiclePayloadToProto(&body.VehicleInfo),
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *DriverHandler) GetDriver(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.GetDriver(ctx, &driverv1.GetDriverRequest{UserId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func vehiclePayloadToProto(body *vehicleInfoPayload) *driverv1.VehicleInfo {
	return &driverv1.VehicleInfo{
		Make:              body.Make,
		Model:             body.Model,
		Color:             body.Color,
		LicensePlate:      body.LicensePlate,
		Seats:             body.Seats,
		Category:          body.Category,
		AcceptsPets:       body.AcceptsPets,
		AcceptsWheelchair: body.AcceptsWheelchair,
		AdditionalInfo:    body.AdditionalInfo,
	}
}
