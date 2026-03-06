package service

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	valueobjects "github.com/nepeta70/ride-hailing/internal/pkg/core/value_objects"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

const (
	// 30 km/h -> 500 m/min
	averageCitySpeedMetersPerMin = 500.0
)

type DirectionsEstimator struct {
}

// Fallback when google maps is not available
// Distance calculates the Haversine distance between two geographic coordinates in meters.
func (dc *DirectionsEstimator) Distance(p1, p2 *core.Coordinates) float64 {
	const earthRadius = 6371230.0 // Radius of the earth in meters

	d2rad := math.Pi / 180

	phi1 := p1.Latitude * d2rad
	phi2 := p2.Latitude * d2rad
	dphi := (p2.Latitude - p1.Latitude) * d2rad
	dlambda := (p2.Longitude - p1.Longitude) * d2rad

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

func (dc *DirectionsEstimator) GetDirections(ctx context.Context, origin, destination *core.Coordinates) (*domain.DirectionsResponse, error) {
	meters := dc.Distance(
		origin,
		destination,
	)

	distance := meters * getCircuityFactor(meters)

	durationMinutes := time.Duration(distance/averageCitySpeedMetersPerMin) * time.Minute

	return &domain.DirectionsResponse{
		DistanceMeters:    distance,
		DurationMinutes:   durationMinutes,
		DurationInTraffic: durationMinutes,
		ArrivalTime:       time.Now().UTC().Add(durationMinutes),
	}, nil
}

func getCircuityFactor(distanceMeters float64) float64 {
	// Short trips are winding (high detour index)
	if distanceMeters < 5000.0 {
		return 1.45
	}
	// Medium trips stabilize
	if distanceMeters < 15000.0 {
		return 1.34
	}
	// Long trips use highways (more direct)
	return 1.25
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
