package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/middleware"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
)

type RideHandler struct {
	client      ridev1.RideServiceClient
	apiKey      string
	hmacSecret  string
	telemetry   ports.TelemetryProvider
	propagator  propagation.TextMapPropagator
}

func NewRideHandler(clients *grpcClients.Clients, cfg *config.Config, telemetry ports.TelemetryProvider) *RideHandler {
	return &RideHandler{
		client:     clients.Ride,
		apiKey:     cfg.Services.Ride.APIKey,
		hmacSecret: cfg.HMACSecret,
		telemetry:  telemetry,
		propagator: telemetry.Propagator(),
	}
}

type coordinatesPayload struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type estimateFareRequest struct {
	PickupLocation  coordinatesPayload `json:"pickup_location" binding:"required"`
	DropoffLocation coordinatesPayload `json:"dropoff_location" binding:"required"`
}

func (h *RideHandler) EstimateFare(c *gin.Context) {
	var body estimateFareRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.EstimateFare(ctx, &ridev1.FareEstimateRequest{
		PickupLocation: &ridev1.Coordinates{
			Latitude:  body.PickupLocation.Latitude,
			Longitude: body.PickupLocation.Longitude,
		},
		DropoffLocation: &ridev1.Coordinates{
			Latitude:  body.DropoffLocation.Latitude,
			Longitude: body.DropoffLocation.Longitude,
		},
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

type requestRidePayload struct {
	FareID          string             `json:"fare_id" binding:"required"`
	ServiceType     string             `json:"service_type" binding:"required"`
	PickupLocation  coordinatesPayload `json:"pickup_location" binding:"required"`
	DropoffLocation coordinatesPayload `json:"dropoff_location" binding:"required"`
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	var body requestRidePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.RequestRide(ctx, &ridev1.RideRequest{
		FareId:      body.FareID,
		ServiceType: body.ServiceType,
		PickupLocation: &ridev1.Coordinates{
			Latitude:  body.PickupLocation.Latitude,
			Longitude: body.PickupLocation.Longitude,
		},
		DropoffLocation: &ridev1.Coordinates{
			Latitude:  body.DropoffLocation.Latitude,
			Longitude: body.DropoffLocation.Longitude,
		},
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *RideHandler) CancelRide(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.CancelRide(ctx, &ridev1.CancelRideRequest{RideId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type acceptRejectPayload struct {
	Accept bool `json:"accept"`
}

func (h *RideHandler) AcceptOrRejectRide(c *gin.Context) {
	var body acceptRejectPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.AcceptOrRejectRide(ctx, &ridev1.AcceptOrRejectRideRequest{
		RideId: c.Param("id"),
		Accept: body.Accept,
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RideHandler) StartRide(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.StartRide(ctx, &ridev1.StartRideRequest{RideId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RideHandler) CompleteRide(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	_, err := h.client.CompleteRide(ctx, &ridev1.CompleteRideRequest{RideId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type fareRatePayload struct {
	ID             string  `json:"id"`
	BaseFare       float64 `json:"base_fare" binding:"required"`
	CostPerKm      float64 `json:"cost_per_km" binding:"required"`
	CostPerMinute  float64 `json:"cost_per_minute" binding:"required"`
	MinimumFare    float64 `json:"minimum_fare" binding:"required"`
	Currency       string  `json:"currency" binding:"required"`
	Country        string  `json:"country" binding:"required"`
	ServiceType    string  `json:"service_type" binding:"required"`
}

func (h *RideHandler) CreateFareRate(c *gin.Context) {
	var body fareRatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.CreateFareRate(ctx, fareRatePayloadToProto(&body))
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *RideHandler) GetFareRates(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.GetFareRates(ctx, &ridev1.GetFareRatesRequest{Country: c.Query("country")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RideHandler) UpdateFareRate(c *gin.Context) {
	var body fareRatePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = c.Param("id")

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.hmacSecret, h.propagator)
	resp, err := h.client.UpdateFareRate(ctx, fareRatePayloadToProto(&body))
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func fareRatePayloadToProto(body *fareRatePayload) *ridev1.FareRate {
	return &ridev1.FareRate{
		Id:            body.ID,
		BaseFare:      body.BaseFare,
		CostPerKm:     body.CostPerKm,
		CostPerMinute: body.CostPerMinute,
		MinimumFare:   body.MinimumFare,
		Currency:      body.Currency,
		Country:       body.Country,
		ServiceType:   body.ServiceType,
	}
}
