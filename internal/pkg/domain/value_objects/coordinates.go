package valueobjects

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type Coordinates struct {
	lat float64
	lon float64
}

func NewCoordinates(lat, lon float64) (*Coordinates, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, errors.NewValidationErrorf("invalid coordinates")
	}
	return &Coordinates{lat: lat, lon: lon}, nil
}

func (c Coordinates) Lat() float64 { return c.lat }

func (c Coordinates) Lon() float64 { return c.lon }
