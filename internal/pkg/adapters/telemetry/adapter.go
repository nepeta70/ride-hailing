package telemetry

import (
	"context"

	telem "github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TelemetryProvider struct {
	meterProvider *sdkmetric.MeterProvider
	Metrics       *telem.Metrics // This is your existing business metrics struct
}

func NewTelemetryProvider(ctx context.Context, serviceName string, cfg *TelemetryConfig) (*TelemetryProvider, error) {
	reg := prom.NewRegistry()

	// 1. Setup the gRPC Exporter (pointing to otel-collector:4317)
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OpentelemetryAddress),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// 2. The Link: Create the Producer from your Registry
	producer := prometheus.NewMetricProducer(prometheus.WithGatherer(reg))

	// 3. Attach the Producer to the Reader
	// The Reader handles the "push loop" automatically every cfg.Interval
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(cfg.Interval),
		sdkmetric.WithProducer(producer),
	)

	// 4. Init the Provider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(provider)

	// 5. Initialize your business metrics with the single Registry
	appMetrics := telem.NewMetrics("ride_hailing", serviceName, reg)

	return &TelemetryProvider{
		meterProvider: provider,
		Metrics:       appMetrics,
	}, nil
}

func (p *TelemetryProvider) Shutdown(ctx context.Context) error {
	return p.meterProvider.Shutdown(ctx)
}
