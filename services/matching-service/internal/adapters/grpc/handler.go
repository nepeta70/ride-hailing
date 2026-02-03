package grpc

import (
	"context"

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/services/matching-service/internal/core/service"
)

type MatchingHandler struct {
	matchingv1.UnimplementedMatchingServiceServer
	service *service.MatchingService
}

func NewMatchingHandler(service *service.MatchingService) *MatchingHandler {
	return &MatchingHandler{service: service}
}

func (h *MatchingHandler) FindMatchingDrivers(ctx context.Context, req *matchingv1.MatchingRequest) (*matchingv1.MatchingResponse, error) {
	// TODO: Call the matching service logic here
	return &matchingv1.MatchingResponse{DriverId: "mock-driver-id"}, nil
}
