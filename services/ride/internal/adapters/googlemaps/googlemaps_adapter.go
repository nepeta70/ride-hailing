package googlemaps

import (
	"context"
	"time"

	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"googlemaps.github.io/maps"
)

type GoogleMapsAdapterOptions struct {
	APIKey          string
	FallBackService ports.DirectionsService
	Telemetry       pkgPorts.TelemetryProvider
	ContextManager  *ctxmgr.ContextManager
}

func (opts *GoogleMapsAdapterOptions) Validate() error {
	if opts.FallBackService == nil {
		return errors.NewValidationErrorf("fallback service is required")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	return nil
}

type GoogleMapsAdapter struct {
	client          *maps.Client
	fallbackService ports.DirectionsService
	telemetry       pkgPorts.TelemetryProvider
	ctxManager      *ctxmgr.ContextManager
}

func NewGoogleMapsAdapter(opts *GoogleMapsAdapterOptions) (ports.DirectionsService, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	if opts.APIKey == "" {
		opts.Telemetry.Logger().Warn("Failed to create Google Maps client: API key is empty. Using fallback distance calculator.")
		return opts.FallBackService, nil // Return fallback service if client initialization fails
	}
	client, err := maps.NewClient(maps.WithAPIKey(opts.APIKey))
	if err != nil {
		opts.Telemetry.Logger().Warn("Failed to create Google Maps client: %v. Using fallback distance calculator.", "error", err)
		return opts.FallBackService, nil // Return fallback service if client initialization fails
	}

	return &GoogleMapsAdapter{
		client:          client,
		fallbackService: opts.FallBackService,
		telemetry:       opts.Telemetry,
		ctxManager:      opts.ContextManager,
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

func (g *GoogleMapsAdapter) GetDirections(ctx context.Context, origin, destination *core.Coordinates) (*domain.DirectionsResponse, error) {
	info, _ := g.ctxManager.Extract(ctx)

	tracer := g.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, "Google Maps GetDirections",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sender.id", info.Sender.ID.String()),
			attribute.String("request.id", info.Trace.RequestID.String()),
			attribute.String("origin", origin.String()),
			attribute.String("destination", destination.String()),
		),
	)
	defer span.End()

	routes, _, err := g.client.Directions(ctx, &maps.DirectionsRequest{
		Origin:        origin.String(),
		Destination:   destination.String(),
		Mode:          maps.TravelModeDriving,
		DepartureTime: "now",
	})

	if err != nil {
		g.telemetry.Logger().WarnContext(ctx, "Google Maps API error: Using fallback distance calculator.", "error", err)
		return g.fallbackService.GetDirections(ctx, origin, destination)
	}
	if len(routes) == 0 || len(routes[0].Legs) == 0 {
		return nil, nil
	}
	leg := routes[0].Legs[0]

	response := &domain.DirectionsResponse{
		DistanceMeters:    float64(leg.Distance.Meters),
		DurationMinutes:   leg.Duration,
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

var _ ports.DirectionsService = (*GoogleMapsAdapter)(nil)
