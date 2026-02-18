package pubsub

import (
	"context"
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"

	"github.com/segmentio/kafka-go"
)

type KafkaSubscriber struct {
	config  *KafkaConfig
	groupID string
	logger  ports.Logger
	readers []*kafka.Reader
	mu      sync.Mutex // Protects the readers slice
	retrier ports.RetrierInterface
}

func NewKafkaSubscriber(config *KafkaConfig, groupID string, retrier ports.RetrierInterface, logger ports.Logger) *KafkaSubscriber {
	return &KafkaSubscriber{
		config:  config,
		groupID: groupID,
		logger:  logger,
		retrier: retrier,
	}
}

func (s *KafkaSubscriber) Subscribe(ctx context.Context, topic string, handler ports.MessageHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  s.config.Brokers,
		GroupID:  s.groupID,
		Topic:    topic,
		MinBytes: s.config.MinBytes,
		MaxBytes: s.config.MaxBytes,
		MaxWait:  s.config.MaxWait,
	})

	s.mu.Lock()
	s.readers = append(s.readers, reader)
	s.mu.Unlock()

	go func() {
		s.logger.Info("Subscriber started", "topic", topic)
		for {
			s.logger.Debug("waiting for message", "topic", topic)
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					s.logger.Info("subscriber context cancelled", "topic", topic)
					return
				}
				s.logger.Error("failed to read message", "error", err, "topic", topic)
				continue
			}

			s.logger.Debug("received kafka message", "topic", topic, "partition", m.Partition, "offset", m.Offset, "size", len(m.Value))

			// Process message with retry logic
			err = s.retrier.Do(ctx, func() error {
				return handler(ctx, m.Value)
			})

			if err != nil {
				// After retries exhausted, check error category
				if errors.IsTransient(err) {
					s.logger.Error("handler failed after retries (transient) - message lost",
						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
					// TODO: Send to DLQ for manual processing
				} else if errors.IsPermanent(err) || errors.IsBusiness(err) {
					s.logger.Warn("handler failed with permanent/business error - skipping message",
						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
					// Don't retry, just log and continue
				} else {
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
			s.logger.Error("failed to close reader", "error", err)
		}
	}
	return nil
}

func (s *KafkaSubscriber) IsHealthy(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.readers {
		_, err := r.FetchMessage(ctx)
		if err != nil {
			s.logger.Error("health check failed for reader", "error", err)
			return false
		}
	}
	return true
}
