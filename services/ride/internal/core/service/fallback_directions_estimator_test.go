package service_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	valueobjects "github.com/nepeta70/ride-hailing/internal/pkg/domain/value_objects"
	. "github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
)

func TestDirectionsEstimator_Distance(t *testing.T) {
	t.Parallel()
	dc := NewDirectionsEstimator()

	tests := []struct {
		name   string
		p1, p2 valueobjects.Coordinates
		want   float64
	}{
		{
			name: "Zero distance",
			p1:   *mustCoordinates(0, 0),
			p2:   *mustCoordinates(0, 0),
			want: 0,
		},
		{
			name: "Known distance",
			p1:   *mustCoordinates(52.5200, 13.4050), // Berlin
			p2:   *mustCoordinates(48.8566, 2.3522),  // Paris
			want: 878000,                             // ~878km
		},
		{
			name: "Antipodal points",
			p1:   *mustCoordinates(0, 0),
			p2:   *mustCoordinates(0, 180),
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

func TestDirectionsEstimator_GetDirections(t *testing.T) {
	dc := NewDirectionsEstimator()

	tests := []struct {
		name        string
		origin      string
		destination string
		wantErr     bool
	}{
		{
			name:        "Valid coordinates",
			origin:      "52.5200,13.4050",
			destination: "48.8566,2.3522",
			wantErr:     false,
		},
		{
			name:        "Invalid origin format",
			origin:      "invalid",
			destination: "48.8566,2.3522",
			wantErr:     true,
		},
		{
			name:        "Invalid destination format",
			origin:      "52.5200,13.4050",
			destination: "invalid",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := dc.GetDirections(context.Background(), tt.origin, tt.destination)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Greater(t, resp.Distance, 0.0)
				assert.Greater(t, resp.Duration, time.Duration(0))
				assert.WithinDuration(t, time.Now().Add(resp.Duration), resp.ArrivalTime, time.Second)
			}
		})
	}
}

func TestCoordinatesFromString(t *testing.T) {

	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLon float64
		wantErr bool
	}{
		{
			name:    "Valid coordinates",
			input:   "52.5200,13.4050",
			wantLat: 52.5200,
			wantLon: 13.4050,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "Non-numeric",
			input:   "abc,def",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := CoordinatesFromString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, coord)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, coord)
				assert.InDelta(t, tt.wantLat, coord.Lat(), 0.0001)
				assert.InDelta(t, tt.wantLon, coord.Lon(), 0.0001)
			}
		})
	}
}

// mustCoordinates is a helper for test setup
func mustCoordinates(lat, lon float64) *valueobjects.Coordinates {
	coord, err := valueobjects.NewCoordinates(lat, lon)
	if err != nil {
		panic(err)
	}
	return coord
}
