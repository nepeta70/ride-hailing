package telemetry

import (
	"context"
	"log/slog"

	"github.com/grafana/pyroscope-go"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	telem "github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	tracer "go.opentelemetry.io/otel/trace"
)

type TelemetryProvider struct {
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	tracerProvider *trace.TracerProvider
	metrics        *telem.Metrics
	tracer         tracer.Tracer
	propagator     propagation.TextMapPropagator
	logger         ports.Logger
}

func NewTelemetryProvider(ctx context.Context, config *config.BaseConfig) (*TelemetryProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	logger, lProvider, err := setupLogger(ctx, config, res)
	if err != nil {
		return nil, err
	}

	metrics, mProvider, err := setupMetrics(ctx, config, res, logger)
	if err != nil {
		return nil, err
	}

	propagator, tracerProvider, err := setupTracing(ctx, config, res)
	if err != nil {
		return nil, err
	}

	err = setupProfiling(config)
	if err != nil {
		logger.Error("Failed to setup profiling:", "error", err)
	}

	return &TelemetryProvider{
		meterProvider:  mProvider,
		loggerProvider: lProvider,
		tracerProvider: tracerProvider,
		metrics:        metrics,
		propagator:     propagator,
		tracer:         tracerProvider.Tracer(config.ServiceName),
		logger:         logger}, nil
}

func (p *TelemetryProvider) Shutdown(ctx context.Context) error {
	if err := p.meterProvider.Shutdown(ctx); err != nil {
		p.logger.Error("Error shutting down MeterProvider", "error", err)
	}
	if err := p.loggerProvider.Shutdown(ctx); err != nil {
		p.logger.Error("Error shutting down LoggerProvider", "error", err)
	}
	if err := p.tracerProvider.Shutdown(ctx); err != nil {
		p.logger.Error("Error shutting down TracerProvider", "error", err)
	}
	return nil
}

func (p *TelemetryProvider) Metrics() ports.Metrics {
	return p.metrics
}

func (p *TelemetryProvider) Logger() ports.Logger {
	return p.logger
}

func (p *TelemetryProvider) Tracer() tracer.Tracer {
	return p.tracer
}

func (p *TelemetryProvider) Propagator() propagation.TextMapPropagator {
	return p.propagator
}

func (p *TelemetryProvider) TracerProvider() *trace.TracerProvider {
	return p.tracerProvider
}

func (p *TelemetryProvider) MeterProvider() *sdkmetric.MeterProvider {
	return p.meterProvider
}

func (p *TelemetryProvider) LoggerProvider() *sdklog.LoggerProvider {
	return p.loggerProvider
}

func setupLogger(ctx context.Context, config *config.BaseConfig, res *resource.Resource) (ports.Logger, *sdklog.LoggerProvider, error) {
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

	otelHandler := otelslog.NewHandler(config.ServiceName,
		otelslog.WithLoggerProvider(lProvider),
	)

	consoleHandler := config.Logging.GetConsoleLogger()

	logHandler := slog.NewMultiHandler(consoleHandler, otelHandler)

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logger, lProvider, nil
}

func setupMetrics(ctx context.Context, config *config.BaseConfig, res *resource.Resource, logger ports.Logger) (*telem.Metrics, *sdkmetric.MeterProvider, error) {
	reg := prom.NewRegistry()
	// 1. Setup the gRPC Exporter
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

func setupTracing(ctx context.Context, config *config.BaseConfig, res *resource.Resource) (propagation.TextMapPropagator, *trace.TracerProvider, error) {
	// Setup the OTLP gRPC exporter for traces
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.Telemetry.OpentelemetryAddress),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)
	return propagator, tp, nil
}

func setupProfiling(config *config.BaseConfig) error {
	pyroscope.Start(pyroscope.Config{
		ApplicationName: config.ServiceName,
		ServerAddress:   config.Telemetry.PyroscopeAddress,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
	return nil
}

var _ ports.TelemetryProvider = (*TelemetryProvider)(nil)
