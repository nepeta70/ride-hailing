package kafka

// import (
// 	"context"
// 	"log/slog"
// 	"sync"
// 	"time"

// 	"github.com/segmentio/kafka-go"
// )

// type KafkaSubscriber struct {
// 	config        *config.KafkaConfig
// 	groupID       string
// 	logger        *slog.Logger
// 	readers       []*kafka.Reader
// 	mu            sync.Mutex // Protects the readers slice
// 	retryStrategy retry.RetryStrategy
// }

// func NewKafkaSubscriber(config *config.KafkaConfig, groupID string, logger *slog.Logger) *KafkaSubscriber {
// 	return &KafkaSubscriber{
// 		config:        config,
// 		groupID:       groupID,
// 		logger:        logger,
// 		retryStrategy: retry.NewExponentialBackoff(retry.DefaultConfig()),
// 	}
// }

// func (s *KafkaSubscriber) Subscribe(ctx context.Context, topic string, handler domain.MessageHandler) error {
// 	reader := kafka.NewReader(kafka.ReaderConfig{
// 		Brokers:  s.config.Brokers,
// 		GroupID:  s.groupID,
// 		Topic:    topic,
// 		MinBytes: 100,
// 		MaxBytes: 10e6,
// 		MaxWait:  100 * time.Millisecond,
// 	})

// 	s.mu.Lock()
// 	s.readers = append(s.readers, reader)
// 	s.mu.Unlock()

// 	// Run the loop you wrote in a goroutine so it doesn't block main
// 	go func() {
// 		s.logger.Info("Subscriber started", "topic", topic)
// 		for {
// 			s.logger.Debug("waiting for message", "topic", topic)
// 			m, err := reader.ReadMessage(ctx)
// 			if err != nil {
// 				if ctx.Err() != nil {
// 					s.logger.Info("subscriber context cancelled", "topic", topic)
// 					return
// 				}
// 				s.logger.Error("failed to read message", "error", err, "topic", topic)
// 				continue
// 			}

// 			s.logger.Debug("received kafka message", "topic", topic, "partition", m.Partition, "offset", m.Offset, "size", len(m.Value))

// 			// Process message with retry logic
// 			err = retry.Do(ctx, s.retryStrategy, func() error {
// 				return handler(ctx, m.Value)
// 			})

// 			if err != nil {
// 				// After retries exhausted, check error category
// 				if domain.IsTransient(err) {
// 					s.logger.Error("handler failed after retries (transient) - message lost",
// 						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
// 					// TODO: Send to DLQ for manual processing
// 				} else if domain.IsPermanent(err) || domain.IsBusiness(err) {
// 					s.logger.Warn("handler failed with permanent/business error - skipping message",
// 						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
// 					// Don't retry, just log and continue
// 				} else {
// 					s.logger.Error("handler failed with unknown error",
// 						"error", err, "topic", topic, "partition", m.Partition, "offset", m.Offset)
// 				}
// 			}
// 		}
// 	}()

// 	return nil
// }

// func (s *KafkaSubscriber) Close() error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	for _, r := range s.readers {
// 		if err := r.Close(); err != nil {
// 			s.logger.Error("failed to close reader", "error", err)
// 		}
// 	}
// 	return nil
// }

// func (s *KafkaSubscriber) IsHealthy(ctx context.Context) bool {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	for _, r := range s.readers {
// 		_, err := r.FetchMessage(ctx)
// 		if err != nil {
// 			s.logger.Error("health check failed for reader", "error", err)
// 			return false
// 		}
// 	}
// 	return true
// }
