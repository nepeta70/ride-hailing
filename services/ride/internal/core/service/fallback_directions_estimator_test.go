package service_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	. "github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
)

func TestDirectionsEstimator_Distance(t *testing.T) {
	t.Parallel()
	dc := NewDirectionsEstimator()

	tests := []struct {
		name   string
		p1, p2 *core.Coordinates
		want   float64
	}{
		{
			name: "Zero distance",
			p1:   mustCoordinates(0, 0),
			p2:   mustCoordinates(0, 0),
			want: 0,
		},
		{
			name: "Known distance",
			p1:   mustCoordinates(52.5200, 13.4050), // Berlin
			p2:   mustCoordinates(48.8566, 2.3522),  // Paris
			want: 878000,                            // ~878km
		},
		{
			name: "Antipodal points",
			p1:   mustCoordinates(0, 0),
			p2:   mustCoordinates(0, 180),
			want: math.Pi * 6371230.0, // half circumference
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dc.Distance(tt.p1, tt.p2)
			assert.InDelta(t, tt.want, got, 1000) // 1km tolerance
		})
	}
}

// mustCoordinates is a helper for test setup
func mustCoordinates(lat, lon float64) *core.Coordinates {
	return &core.Coordinates{
		Latitude:  lat,
		Longitude: lon,
	}
}
