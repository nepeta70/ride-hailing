package grpc

import (
	"context"

	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/app"
)

type MatchingHandler struct {
	matchingv1.UnimplementedMatchingServiceServer
	app *app.Application
}

func NewMatchingHandler(app *app.Application) *MatchingHandler {
	return &MatchingHandler{app: app}
}

func (h *MatchingHandler) FindMatchingDrivers(ctx context.Context, req *matchingv1.MatchingRequest) (*matchingv1.MatchingResponse, error) {
	// TODO: Call the matching service logic here
	return &matchingv1.MatchingResponse{DriverId: "mock-driver-id"}, nil
}
