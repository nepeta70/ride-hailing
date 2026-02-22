package telemetry

import (
	"context"
	"log/slog"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	telem "github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TelemetryProvider struct {
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	metrics        *telem.Metrics // This is your existing business metrics struct
	logger         ports.Logger
}

func NewTelemetryProvider(ctx context.Context, config *config.BaseConfig) (*TelemetryProvider, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	logger, lProvider, err := SetupLogger(ctx, config, res)
	if err != nil {
		return nil, err
	}

	metrics, mProvider, err := SetupMetrics(ctx, config, res, logger)
	if err != nil {
		return nil, err
	}

	return &TelemetryProvider{
		meterProvider:  mProvider,
		loggerProvider: lProvider,
		metrics:        metrics,
		logger:         logger,
	}, nil
}

func (p *TelemetryProvider) Shutdown(ctx context.Context) error {
	if err := p.meterProvider.Shutdown(ctx); err != nil {
		p.logger.Error("Error shutting down MeterProvider", "error", err)
	}
	if err := p.loggerProvider.Shutdown(ctx); err != nil {
		p.logger.Error("Error shutting down LoggerProvider", "error", err)
	}
	return nil
}

func (p *TelemetryProvider) GetMetrics() *telem.Metrics {
	return p.metrics
}

func (p *TelemetryProvider) GetLogger() ports.Logger {
	return p.logger
}

func SetupLogger(ctx context.Context, config *config.BaseConfig, res *resource.Resource) (ports.Logger, *sdklog.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(config.Telemetry.OpentelemetryAddress),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	lProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)

	// Register as global so bridges (Zap/Slog) can find it
	otelHandler := otelslog.NewHandler(config.ServiceName,
		otelslog.WithLoggerProvider(lProvider),
	)

	consoleHandler := config.Logging.GetConsoleLogger()

	logHandler := slog.NewMultiHandler(consoleHandler, otelHandler)

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logger, lProvider, nil
}

func SetupMetrics(ctx context.Context, config *config.BaseConfig, res *resource.Resource, logger ports.Logger) (*telem.Metrics, *sdkmetric.MeterProvider, error) {
	reg := prom.NewRegistry()
	// 1. Setup the gRPC Exporter (pointing to otel-collector:4317)
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(config.Telemetry.OpentelemetryAddress),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		logger.Error("Failed to create OTLP gRPC exporter:", "error", err)
		return nil, nil, err
	}

	// 2. The Link: Create the Producer from your Registry
	producer := prometheus.NewMetricProducer(prometheus.WithGatherer(reg))

	// 3. Attach the Producer to the Reader
	// The Reader handles the "push loop" automatically every cfg.Interval
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(config.Telemetry.Interval),
		sdkmetric.WithProducer(producer),
	)

	// 4. Init the Provider
	mProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(mProvider)

	// 5. Initialize your business metrics with the single Registry
	appMetrics := telem.NewMetrics("ride_hailing", config.ServiceName, reg)
	return appMetrics, mProvider, nil
}
