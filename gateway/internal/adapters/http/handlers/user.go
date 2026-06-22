package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	grpcClients "github.com/nepeta70/ride-hailing/gateway/internal/adapters/grpc"
	"github.com/nepeta70/ride-hailing/gateway/internal/adapters/http/middleware"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
)

type UserHandler struct {
	client     userv1.UserServiceClient
	apiKey     string
	propagator propagation.TextMapPropagator
}

func NewUserHandler(clients *grpcClients.Clients, cfg *config.Config, telemetry ports.TelemetryProvider) *UserHandler {
	return &UserHandler{
		client:     clients.User,
		apiKey:     cfg.Services.User.APIKey,
		propagator: telemetry.Propagator(),
	}
}

type createUserPayload struct {
	UserType    string `json:"user_type" binding:"required"`
	UserName    string `json:"user_name" binding:"required"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var body createUserPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.propagator)
	resp, err := h.client.CreateUser(ctx, &userv1.CreateUserRequest{
		UserType:    body.UserType,
		UserName:    body.UserName,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Password:    body.Password,
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.propagator)
	resp, err := h.client.GetUser(ctx, &userv1.GetUserRequest{UserId: c.Param("id")})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

type updateUserPayload struct {
	UserName    string `json:"user_name"`
	UserType    string `json:"user_type"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var body updateUserPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := middleware.OutgoingGRPCContext(c, h.apiKey, h.propagator)
	resp, err := h.client.UpdateUser(ctx, &userv1.UpdateUserRequest{
		UserId:      c.Param("id"),
		UserName:    body.UserName,
		UserType:    body.UserType,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Password:    body.Password,
	})
	if err != nil {
		middleware.WriteGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
