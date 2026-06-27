package mongodb

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const mongoAdapterServiceName = "MongoAdapter"

type MongoAdapterOpts struct {
	Config         *MongoConfig
	RetrierFactory ports.RetrierFactoryInterface
	Telemetry      ports.TelemetryProvider
}

func (m *MongoAdapterOpts) Validate() error {
	if m.Config == nil {
		return errors.NewValidationErrorf("Config is required")
	}

	if m.RetrierFactory == nil {
		return errors.NewValidationErrorf("RetryFactory is required")
	}

	if m.Telemetry == nil {
		return errors.NewValidationErrorf("Telemetry is required")
	}

	return nil
}

type MongoAdapter struct {
	cfg          *MongoConfig
	retryFactory ports.RetrierFactoryInterface
	telemetry    ports.TelemetryProvider
	Client       *mongo.Client
}

func NewMongoAdapter(opts *MongoAdapterOpts, ctx context.Context) (*MongoAdapter, error) {
	ctx, span := opts.Telemetry.Tracer().Start(ctx, "MongoAdapter:Initialize",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.Bool("service.init", true)),
	)
	defer span.End()

	err := opts.Validate()
	if err != nil {
		opts.Telemetry.Logger().ErrorContext(ctx, "Failed to validate mongo adapter", "error", err)
		span.SetStatus(codes.Error, "failed to validate mongo adapter")
		return nil, errors.NewPermanentErrorf("failed to validate mongo adapter: %w", err)
	}

	uri := opts.Config.GetURI()

	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(opts.Config.PingTimeout)
	clientOptions.SetServerSelectionTimeout(opts.Config.PingTimeout)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, errors.NewPermanentErrorf("failed to create mongo client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, opts.Config.PingTimeout)
	defer cancel()

	retrier := opts.RetrierFactory.NewExponentialBackoffRetrier(mongoAdapterServiceName, opts.Config.PingTimeout)
	err = retrier.Do(ctx, func(ctx context.Context) error {
		return client.Ping(pingCtx, readpref.Primary())
	})

	if err != nil {
		return nil, errors.NewPermanentErrorf("failed to ping mongodb after retries: %w", err)
	}

	return &MongoAdapter{
		cfg:          opts.Config,
		retryFactory: opts.RetrierFactory,
		telemetry:    opts.Telemetry,
		Client:       client,
	}, nil
}

func (m *MongoAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.cfg.PingTimeout)
	defer cancel()
	if err := m.Client.Ping(ctx, readpref.Primary()); err != nil {
		return errors.NewTransientErrorf("failed to ping mongodb: %w", err)
	}
	return nil
}

func (m *MongoAdapter) ServiceName() string {
	return mongoAdapterServiceName
}

func (m *MongoAdapter) Close() error {
	return m.Client.Disconnect(context.Background())
}
