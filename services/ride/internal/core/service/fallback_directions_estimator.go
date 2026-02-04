package service

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	valueobjects "github.com/nepeta70/ride-hailing/internal/pkg/domain/value_objects"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type DirectionsEstimator struct {
}

// Fallback when google maps is not available
// Distance calculates the Haversine distance between two geographic coordinates in meters.
func (dc *DirectionsEstimator) Distance(p1, p2 valueobjects.Coordinates) float64 {
	const earthRadius = 6371230.0 // Radius of the earth in meters

	d2rad := math.Pi / 180

	phi1 := p1.Lat() * d2rad
	phi2 := p2.Lat() * d2rad
	dphi := (p2.Lat() - p1.Lat()) * d2rad
	dlambda := (p2.Lon() - p1.Lon()) * d2rad

	sindphi2 := math.Sin(dphi / 2)
	sindlambda2 := math.Sin(dlambda / 2)

	a := sindphi2*sindphi2 +
		math.Cos(phi1)*math.Cos(phi2)*
			sindlambda2*sindlambda2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func NewDirectionsEstimator() *DirectionsEstimator {
	return &DirectionsEstimator{}
}

func (dc *DirectionsEstimator) GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error) {
	// This is a fallback implementation and does not provide real directions.
	// In a real-world scenario, you would call an external service here.
	pickupCoords, err := CoordinatesFromString(origin)
	if err != nil {
		return nil, errors.BusinessError(err)
	}
	dropoffCoords, err := CoordinatesFromString(destination)
	if err != nil {
		return nil, errors.BusinessError(err)
	}
	a := dc.Distance(
		*pickupCoords,
		*dropoffCoords,
	)

	duration := time.Duration((a/40000)*3600) * time.Second
	return &domain.DirectionsResponse{
		Distance:          a,
		Duration:          duration,
		DurationInTraffic: duration,
		ArrivalTime:       time.Now().Add(duration),
	}, nil
}

func CoordinatesFromString(s string) (*valueobjects.Coordinates, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return nil, errors.NewValidationErrorf("coordinates must be in 'lat, lon' format")
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil, errors.NewValidationErrorf("invalid latitude")
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, errors.NewValidationErrorf("invalid longitude")
	}

	coordinates, err := valueobjects.NewCoordinates(lat, lon)
	if err != nil {
		return nil, err
	}
	return coordinates, nil
}
