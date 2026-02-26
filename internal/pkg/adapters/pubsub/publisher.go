package pubsub

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type KafkaPublisherOptions struct {
	Config         *KafkaConfig
	TopicProvider  ports.TopicProvider
	Telemetry      ports.TelemetryProvider
	RetrierFactory ports.RetrierFactoryInterface
	ContextManager *ctxmgr.ContextManager
}

func (o *KafkaPublisherOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("KafkaConfig is required")
	}
	if o.TopicProvider == nil {
		return errors.NewValidationErrorf("TopicProvider is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("TelemetryProvider is required")
	}
	if o.RetrierFactory == nil {
		return errors.NewValidationErrorf("RetrierFactory is required")
	}
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("ContextManager is required")
	}
	return nil
}

type KafkaPublisher struct {
	topicProvider  ports.TopicProvider
	writer         *kafka.Writer
	brokers        []string
	telemetry      ports.TelemetryProvider
	retrierFactory ports.RetrierFactoryInterface
	contextManager *ctxmgr.ContextManager
}

func NewEventPublisher(opts *KafkaPublisherOptions) (ports.EventPublisher, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	kp := &KafkaPublisher{
		brokers:        opts.Config.Brokers,
		topicProvider:  opts.TopicProvider,
		telemetry:      opts.Telemetry,
		retrierFactory: opts.RetrierFactory,
		contextManager: opts.ContextManager,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(opts.Config.Brokers...),
			Balancer:     &kafka.LeastBytes{},
			Async:        true,
			RequiredAcks: kafka.RequireOne,
			BatchSize:    opts.Config.BatchSize,
			BatchTimeout: time.Duration(opts.Config.BatchTimeoutMs) * time.Millisecond,
			ErrorLogger:  &kafkaErrorLogger{logger: opts.Telemetry.Logger()},
		},
	}

	if opts.Config.EnableLogging {
		kp.writer.Logger = &kafkaLogger{logger: opts.Telemetry.Logger()}
	}

	if opts.Config.AutoCreate {
		err := kp.initializeTopics(opts.TopicProvider.AllTopics())
		if err != nil {
			return nil, err
		}
	}

	return kp, nil
}

func (k *KafkaPublisher) Publish(ctx context.Context, topic contracts.Topic, message *contracts.EventMessage) error {
	tracer := k.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, "Kafka Publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("topic", topic.String()),
			attribute.String("operation", "publish"),
			attribute.String("message_type", message.EventType.String()),
		),
	)

	defer span.End()

	data, err := json.Marshal(message)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	carrier := propagation.MapCarrier{}
	maps.Copy(carrier, message.Headers)
	k.telemetry.Propagator().Inject(ctx, carrier)

	k.telemetry.Logger().Debug("publishing message", "message", message.Payload)
	k.telemetry.Logger().Debug("message headers", "headers", message.Headers)
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic.String(),
		Value:   data,
		Headers: k.makeHeader(message),
		Time:    time.Now().UTC(),
	})
	if err != nil {
		k.telemetry.Metrics().DependencyFailure(k.ServiceName(), "publish", "error")
		return errors.NewTransientErrorf("failed to publish message: %v", err)
	}
	return nil
}

func (k *KafkaPublisher) HealthCheck(ctx context.Context) error {
	conn, err := k.dial(ctx)
	if err != nil {
		k.telemetry.Metrics().DependencyFailure(k.ServiceName(), "dial", "error")
		return errors.NewTransientErrorf("Kafka connection failed: %v", err)
	}
	defer conn.Close()

	_, err = conn.ReadPartitions()
	if err != nil {
		k.telemetry.Metrics().DependencyFailure(k.ServiceName(), "read_partitions", "error")
		return errors.NewTransientErrorf("Kafka health check failed: %v", err)
	}
	return nil
}

func (k *KafkaPublisher) EnsureTopics(topics ...string) error {
	conn, err := k.dial(context.Background())
	if err != nil {
		k.telemetry.Metrics().DependencyFailure(k.ServiceName(), "dial", "error")
		return errors.NewTransientErrorf("Kafka connection failed: %v", err)
	}
	defer conn.Close()

	configs := make([]kafka.TopicConfig, len(topics))
	for i, t := range topics {
		configs[i] = kafka.TopicConfig{Topic: t, NumPartitions: 1, ReplicationFactor: 1}
	}

	return conn.CreateTopics(configs...)
}

func (k *KafkaPublisher) Close() error {
	return k.writer.Close()
}

func (k *KafkaPublisher) TopicProvider() ports.TopicProvider {
	return k.topicProvider
}

func (k *KafkaPublisher) ServiceName() string {
	return "kafka-publisher"
}

func (k *KafkaPublisher) initializeTopics(required []string) error {
	retrier := k.retrierFactory.NewExponentialBackoffRetrier(k.ServiceName(), 30*time.Second)
	err := retrier.Do(context.Background(), func() error {
		if err := k.EnsureTopics(required...); err != nil {
			k.telemetry.Logger().Error("Failed to create Kafka topics, will retry", "topics", required, "error", err)
		} else if k.verify(required) {
			k.telemetry.Logger().Info("Kafka Ready: All topics verified", "topics", required)
			return nil
		}
		return errors.NewTransientErrorf("Failed to ensure Kafka topics")
	})
	if err != nil {
		k.telemetry.Logger().Error("Failed to ensure Kafka topics after retries", "topics", required, "error", err)
		k.telemetry.Metrics().DependencyFailure(k.ServiceName(), "initialize_topics", "error")
		return errors.NewPermanentErrorf("Failed to initialize Kafka topics: %v", err)
	}

	return nil
}

// Private helper to avoid repeating dial logic
func (k *KafkaPublisher) dial(ctx context.Context) (*kafka.Conn, error) {
	return kafka.DialContext(ctx, "tcp", k.brokers[0])
}

func (k *KafkaPublisher) verify(required []string) bool {
	conn, err := k.dial(context.Background())
	if err != nil {
		return false
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false
	}

	existing := make(map[string]bool)
	for _, p := range partitions {
		existing[p.Topic] = true
	}

	for _, r := range required {
		if !existing[r] {
			return false
		}
	}
	return true
}

func (k *KafkaPublisher) makeHeader(message *contracts.EventMessage) []kafka.Header {
	headers := make([]kafka.Header, 0, len(message.Headers))
	for k, v := range message.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	return headers
}

var _ ports.EventPublisher = (*KafkaPublisher)(nil)
var _ ports.HealthProvider = (*KafkaPublisher)(nil)
