package domain

import (
	"fmt"
	"strconv"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (c *Coordinates) String() string {
	if c == nil {
		return "(0, 0)"
	}

	b := make([]byte, 0, 40)
	b = append(b, '(')
	b = strconv.AppendFloat(b, c.Latitude, 'f', 6, 64)
	b = append(b, ',', ' ')
	b = strconv.AppendFloat(b, c.Longitude, 'f', 6, 64)
	b = append(b, ')')

	return string(b)
}

func (c *Coordinates) Validate() error {
	// Latitude range: [-90, 90]
	if c.Latitude < -90 || c.Latitude > 90 {
		return errors.NewValidationErrorf("invalid latitude: %f", c.Latitude)
	}
	// Longitude range: [-180, 180]
	if c.Longitude < -180 || c.Longitude > 180 {
		return errors.NewValidationErrorf("invalid longitude: %f", c.Longitude)
	}
	return nil
}

func ParseCoordinates(coords string) (*Coordinates, error) {
	var c Coordinates
	_, err := fmt.Sscanf(coords, "(%f, %f)", &c.Latitude, &c.Longitude)
	if err != nil {
		return nil, errors.NewValidationErrorf("invalid coordinates format: %s", coords)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func NewCoordinates(lat, lon float64) (*Coordinates, error) {
	c := Coordinates{
		Latitude:  lat,
		Longitude: lon,
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}
