package pubsub

import (
	"context"
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/segmentio/kafka-go"
)

type KafkaSubscriberOptions struct {
	Config         *KafkaConfig
	GroupID        string
	RetrierFactory ports.RetrierFactoryInterface
	Telemetry      ports.TelemetryProvider
}

func (o *KafkaSubscriberOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("KafkaConfig is required")
	}
	if o.GroupID == "" {
		return errors.NewValidationErrorf("GroupID is required")
	}
	if o.RetrierFactory == nil {
		return errors.NewValidationErrorf("RetrierFactory is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("Telemetry is required")
	}
	return nil
}

type KafkaSubscriber struct {
	config         *KafkaConfig
	groupID        string
	readers        []*kafka.Reader
	mu             sync.Mutex // Protects the readers slice
	retrierFactory ports.RetrierFactoryInterface
	telemetry      ports.TelemetryProvider
}

func NewKafkaSubscriber(opts *KafkaSubscriberOptions) (*KafkaSubscriber, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &KafkaSubscriber{
		config:         opts.Config,
		groupID:        opts.GroupID,
		retrierFactory: opts.RetrierFactory,
		telemetry:      opts.Telemetry,
	}, nil
}

func (s *KafkaSubscriber) Subscribe(ctx context.Context, topic contracts.Topic, handler ports.MessageHandler) error {
	var kfLogger kafka.Logger
	if s.config.EnableLogging {
		kfLogger = &kafkaLogger{logger: s.telemetry.Logger()}
	}
	sTopic := topic.String()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           s.config.Brokers,
		GroupID:           s.groupID,
		Topic:             sTopic,
		MinBytes:          s.config.MinBytes,
		MaxBytes:          s.config.MaxBytes,
		MaxWait:           s.config.MaxWait,
		SessionTimeout:    s.config.SessionTimeout,
		RebalanceTimeout:  s.config.RebalanceTimeout,
		HeartbeatInterval: s.config.HeartbeatInterval,
		Logger:            kfLogger,
		ErrorLogger:       &kafkaErrorLogger{logger: s.telemetry.Logger()},
	})

	s.mu.Lock()
	s.readers = append(s.readers, reader)
	s.mu.Unlock()

	go func() {
		s.telemetry.Logger().Debug("Subscriber started", "topic", sTopic)
		for {
			s.telemetry.Logger().Debug("Waiting for message", "topic", sTopic)
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					s.telemetry.Logger().Debug("Subscriber context cancelled", "topic", sTopic)
					return
				}
				s.telemetry.Logger().Error("Failed to read message", "error", err, "topic", sTopic)
				continue
			}

			s.telemetry.Logger().Debug("Received kafka message", "topic", sTopic, "partition", message.Partition, "offset", message.Offset, "size", len(message.Value))

			// Process message with retry logic
			retrier := s.retrierFactory.NewExponentialBackoffRetrier(s.ServiceName(), s.config.BatchTimeout)
			err = retrier.Do(ctx, func() error {
				headers := make(map[string]string)
				for _, h := range message.Headers {
					headers[h.Key] = string(h.Value)
				}
				return handler(ctx, headers, message.Value)
			})

			if err != nil {
				// After retries exhausted, check error category
				if errors.IsTransient(err) {
					s.telemetry.Metrics().DependencyFailure(s.ServiceName(), "message_handler", "transient_error")
					s.telemetry.Logger().Error("handler failed after retries (transient) - message lost",
						"error", err, "topic", sTopic, "partition", message.Partition, "offset", message.Offset)
					// TODO: Send to DLQ for manual processing
				} else if errors.IsPermanent(err) || errors.IsBusiness(err) {
					s.telemetry.Metrics().DependencyFailure(s.ServiceName(), "message_handler", "permanent_error")
					s.telemetry.Logger().Warn("handler failed with permanent/business error - skipping message",
						"error", err, "topic", sTopic, "partition", message.Partition, "offset", message.Offset)
					// Don't retry, just log and continue
				} else {
					s.telemetry.Metrics().DependencyFailure(s.ServiceName(), "message_handler", "unknown_error")
					s.telemetry.Logger().Error("handler failed with unknown error",
						"error", err, "topic", sTopic, "partition", message.Partition, "offset", message.Offset)
				}
			}
		}
	}()

	return nil
}

func (s *KafkaSubscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.readers {
		if err := r.Close(); err != nil {
			s.telemetry.Logger().Error("failed to close kafka reader", "error", err)
			s.telemetry.Metrics().DependencyFailure(s.ServiceName(), "close_reader", "error")
		}
	}
	return nil
}

func (s *KafkaSubscriber) HealthCheck(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.readers {
		_, err := r.FetchMessage(ctx)
		if err != nil {
			s.telemetry.Logger().Error("health check failed for kafka reader", "error", err)
			s.telemetry.Metrics().DependencyFailure(s.ServiceName(), "health_check", "error")
			return errors.NewTransientErrorf("Health check failed for Kafka subscriber: %w", err)
		}
	}
	return nil
}

func (s *KafkaSubscriber) ServiceName() string {
	return "kafka-subscriber"
}

var _ ports.EventSubscriber = (*KafkaSubscriber)(nil)
var _ ports.HealthProvider = (*KafkaSubscriber)(nil)
