package service

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
)

type MatchingService struct {
	client ports.GetCandidates
}

func NewMatchingService(client ports.GetCandidates) *MatchingService {
	return &MatchingService{client: client}
}

func (s *MatchingService) MatchRiderToDriver(ctx context.Context, request *contracts.RideRequestedEvent) (string, error) {
	// candidates, err := s.client.GetCandidates(ctx, request.PickupLocation)
	// if err != nil {
	// 	return "", err
	// }
	// publish event

	// Implement matching logic here
	return "driver-id-placeholder", nil
}
