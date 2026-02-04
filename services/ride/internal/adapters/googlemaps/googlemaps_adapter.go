package googlemaps

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	ridePorts "github.com/nepeta70/ride-hailing/services/ride/internal/ports"

	"googlemaps.github.io/maps"
)

type GoogleMapsAdapter struct {
	client *maps.Client
}

func NewGoogleMapsAdapter(cfg *config.Config) (*GoogleMapsAdapter, error) {
	client, err := maps.NewClient(maps.WithAPIKey(cfg.KeysConfig.GoogleMapsAPIKey))
	if err != nil {
		return nil, err
	}

	return &GoogleMapsAdapter{
		client: client,
	}, nil
}

func (g *GoogleMapsAdapter) HealthCheck(ctx context.Context) error {
	// Google Maps API does not provide a direct health check endpoint.
	// A simple way to check health is to make a lightweight request.
	_, err := g.client.Timezone(ctx, &maps.TimezoneRequest{
		Location:  &maps.LatLng{Lat: 0, Lng: 0},
		Timestamp: time.Now(),
	})
	return err
}
func (g *GoogleMapsAdapter) ServiceName() string {
	return "Google Maps API"
}

func (g *GoogleMapsAdapter) GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error) {
	routes, _, err := g.client.Directions(ctx, &maps.DirectionsRequest{
		Origin:        origin,
		Destination:   destination,
		Mode:          maps.TravelModeDriving,
		DepartureTime: "now",
	})

	if err != nil {
		return nil, err
	}
	if len(routes) == 0 || len(routes[0].Legs) == 0 {
		return nil, nil
	}
	leg := routes[0].Legs[0]

	response := &domain.DirectionsResponse{
		Distance:          float64(leg.Distance.Meters),
		Duration:          leg.Duration,
		DurationInTraffic: leg.DurationInTraffic,
		ArrivalTime:       leg.ArrivalTime,
	}
	return response, nil
}

// 	// DepartureTime contains the estimated time of departure for this leg. This
// 	// property is only returned for transit directions.
// 	DepartureTime time.Time `json:"departure_time"`

// 	// StartLocation contains the latitude/longitude coordinates of the origin of this
// 	// leg.
// 	StartLocation LatLng `json:"start_location"`

// 	// EndLocation contains the latitude/longitude coordinates of the destination of
// 	// this leg.
// 	EndLocation LatLng `json:"end_location"`

// 	// StartAddress contains the human-readable address (typically a street address)
// 	// reflecting the start location of this leg.
// 	StartAddress string `json:"start_address"`

// 	// EndAddress contains the human-readable address (typically a street address)
// 	// reflecting the end location of this leg.
// 	EndAddress string `json:"end_address"`

// 	// ViaWaypoint contains info about points through which the route was laid.
// 	ViaWaypoint []*ViaWaypoint `json:"via_waypoint"`
// }

var _ ridePorts.DirectionsService = (*GoogleMapsAdapter)(nil)
