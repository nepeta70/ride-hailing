package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/middleware"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
)

type LocationHandler struct {
	client      locationv1.LocationServiceClient
	apiKey      string
	hmacSecret  string
	propagator  propagation.TextMapPropagator
}

func NewLocationHandler(clients *grpcClients.Clients, cfg *config.Config, telemetry ports.TelemetryProvider) *LocationHandler {
	return &LocationHandler{
		client:     clients.Location,
		apiKey:     cfg.Services.Location.APIKey,
		hmacSecret: cfg.HMACSecret,
		propagator: telemetry.Propagator(),
	}
}

type driverLocationPayload struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Accuracy  float32 `json:"accuracy"`
	Heading   float32 `json:"heading"`
	Speed     float32 `json:"speed"`
	Status    string  `json:"status"`
}

func (h *LocationHandler) UpdateDriverLocation(c *gin.Context) {
	var body driverLocationPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.UpdateDriverLocation(ctx, &locationv1.DriverLocation{
		Latitude:  body.Latitude,
		Longitude: body.Longitude,
		Accuracy:  body.Accuracy,
		Heading:   body.Heading,
		Speed:     body.Speed,
		Status:    body.Status,
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LocationHandler) GetDriverLocation(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.GetDriverLocation(ctx, &locationv1.UserID{UserId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *LocationHandler) DeleteDriverLocation(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.DeleteDriverLocation(ctx, &locationv1.UserID{UserId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LocationHandler) SearchNearbyDrivers(c *gin.Context) {
	latitude, err := strconv.ParseFloat(c.Query("latitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude is required"})
		return
	}
	longitude, err := strconv.ParseFloat(c.Query("longitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "longitude is required"})
		return
	}
	radiusKm, err := strconv.ParseFloat(c.DefaultQuery("radius_km", "5"), 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius_km"})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.SearchNearbyDrivers(ctx, &locationv1.SearchNearbyDriversRequest{
		Latitude:  latitude,
		Longitude: longitude,
		RadiusKm:  float32(radiusKm),
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

type driverStatusPayload struct {
	Status string `json:"status" binding:"required"`
}

func (h *LocationHandler) UpdateDriverStatus(c *gin.Context) {
	var body driverStatusPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.UpdateDriverStatus(ctx, &locationv1.DriverStatus{
		DriverId: c.Param("id"),
		Status:   body.Status,
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
