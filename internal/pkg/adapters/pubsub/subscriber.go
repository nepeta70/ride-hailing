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
	Logger         ports.Logger
	RetrierFactory ports.RetrierFactoryInterface
	Metrics        ports.Metrics
}

func (o *KafkaSubscriberOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("KafkaConfig is required")
	}
	if o.GroupID == "" {
		return errors.NewValidationErrorf("GroupID is required")
	}
	if o.Logger == nil {
		return errors.NewValidationErrorf("Logger is required")
	}
	if o.RetrierFactory == nil {
		return errors.NewValidationErrorf("RetrierFactory is required")
	}
	if o.Metrics == nil {
		return errors.NewValidationErrorf("Metrics is required")
	}
	return nil
}

type KafkaSubscriber struct {
	config         *KafkaConfig
	groupID        string
	logger         ports.Logger
	readers        []*kafka.Reader
	mu             sync.Mutex // Protects the readers slice
	retrierFactory ports.RetrierFactoryInterface
	metrics        ports.Metrics
}

func NewKafkaSubscriber(opts *KafkaSubscriberOptions) (*KafkaSubscriber, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &KafkaSubscriber{
		config:         opts.Config,
		groupID:        opts.GroupID,
		logger:         opts.Logger,
		retrierFactory: opts.RetrierFactory,
		metrics:        opts.Metrics,
	}, nil
}

func (s *KafkaSubscriber) Subscribe(ctx context.Context, topic contracts.Topic, handler ports.MessageHandler) error {
	var kfLogger kafka.Logger
	if s.config.EnableLogging {
		kfLogger = &kafkaLogger{logger: s.logger}
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     s.config.Brokers,
		GroupID:     s.groupID,
		Topic:       string(topic),
		MinBytes:    s.config.MinBytes,
		MaxBytes:    s.config.MaxBytes,
		MaxWait:     s.config.MaxWait,
		Logger:      kfLogger,
		ErrorLogger: &kafkaErrorLogger{logger: s.logger},
	})

	s.mu.Lock()
	s.readers = append(s.readers, reader)
	s.mu.Unlock()

	go func() {
		s.logger.Debug("Subscriber started", "topic", topic)
		for {
			s.logger.Debug("waiting for message", "topic", topic)
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					s.logger.Debug("subscriber context cancelled", "topic", topic)
					return
				}
				s.logger.Error("failed to read message", "error", err, "topic", topic)
				continue
			}

			s.logger.Debug("received kafka message", "topic", topic, "partition", m.Partition, "offset", m.Offset, "size", len(m.Value))

			// Process message with retry logic
			retrier := s.retrierFactory.NewExponentialBackoffRetrier(s.ServiceName(), s.config.BatchTimeout)
			err = retrier.Do(ctx, func() error {
				return handler(ctx, m.Value)
			})

			if err != nil {
				// After retries exhausted, check error category
				if errors.IsTransient(err) {
					s.metrics.DependencyFailure(s.ServiceName(), "message_handler", "transient_error")
					s.logger.Error("handler failed after retries (transient) - message lost",
						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
					// TODO: Send to DLQ for manual processing
				} else if errors.IsPermanent(err) || errors.IsBusiness(err) {
					s.metrics.DependencyFailure(s.ServiceName(), "message_handler", "permanent_error")
					s.logger.Warn("handler failed with permanent/business error - skipping message",
						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
					// Don't retry, just log and continue
				} else {
					s.metrics.DependencyFailure(s.ServiceName(), "message_handler", "unknown_error")
					s.logger.Error("handler failed with unknown error",
						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
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
			s.logger.Error("failed to close kafka reader", "error", err)
			s.metrics.DependencyFailure(s.ServiceName(), "close_reader", "error")
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
			s.logger.Error("health check failed for kafka reader", "error", err)
			s.metrics.DependencyFailure(s.ServiceName(), "health_check", "error")
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
