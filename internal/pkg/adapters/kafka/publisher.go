package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"time"

// 	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
// 	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
// 	"github.com/segmentio/kafka-go"
// )

// type KafkaPublisher struct {
// 	topicProvider domain.TopicProvider
// 	writer        *kafka.Writer
// 	brokers       []string
// 	logger        ports.Logger
// }

// func NewEventPublisher(cfg config.KafkaConfig, topicProvider domain.TopicProvider, logger ports.Logger) domain.EventPublisher {
// 	kp := &KafkaPublisher{
// 		brokers:       cfg.Brokers,
// 		topicProvider: topicProvider,
// 		logger:        logger,
// 		writer: &kafka.Writer{
// 			Addr:         kafka.TCP(cfg.Brokers...),
// 			Balancer:     &kafka.LeastBytes{},
// 			Async:        true,
// 			RequiredAcks: kafka.RequireOne,
// 			BatchSize:    cfg.BatchSize,
// 			BatchTimeout: time.Duration(cfg.BatchTimeout) * time.Millisecond,
// 		},
// 	}

// 	if cfg.EnableLogging {
// 		bridge := &kafkaLogger{logger: logger}
// 		kp.writer.Logger = bridge
// 		kp.writer.ErrorLogger = bridge
// 	}

// 	if cfg.AutoCreate {
// 		kp.initializeTopics(topicProvider.AllTopics())
// 	}

// 	return kp
// }

// func (k *KafkaPublisher) Publish(ctx context.Context, topic string, message *domain.EventMessage) error {
// 	data, err := json.Marshal(message)
// 	if err != nil {
// 		return errors.NewErrJSONMarshal(err)
// 	}
// 	return k.writer.WriteMessages(ctx, kafka.Message{
// 		Topic:   topic,
// 		Value:   data,
// 		Headers: []kafka.Header{{Key: "eventType", Value: []byte(message.EventType)}},
// 		Time:    message.Timestamp,
// 	})
// }

// func (k *KafkaPublisher) IsHealthy(ctx context.Context) bool {
// 	conn, err := k.dial(ctx)
// 	if err != nil {
// 		return false
// 	}
// 	defer conn.Close()

// 	_, err = conn.ReadPartitions()
// 	return err == nil
// }

// func (k *KafkaPublisher) EnsureTopics(topics ...string) error {
// 	conn, err := k.dial(context.Background())
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	configs := make([]kafka.TopicConfig, len(topics))
// 	for i, t := range topics {
// 		configs[i] = kafka.TopicConfig{Topic: t, NumPartitions: 1, ReplicationFactor: 1}
// 	}

// 	return conn.CreateTopics(configs...)
// }

// func (k *KafkaPublisher) Close() error {
// 	return k.writer.Close()
// }

// func (k *KafkaPublisher) TopicProvider() domain.TopicProvider {
// 	return k.topicProvider
// }

// func (k *KafkaPublisher) initializeTopics(required []string) {
// 	for i := range 10 {
// 		if err := k.EnsureTopics(required...); err != nil {
// 			k.logger.Warn("Failed to ensure topics", "attempt", i+1, "err", err)
// 		} else if k.verify(required) {
// 			k.logger.Info("Kafka Ready: All topics verified", "topics", required)
// 			return
// 		}
// 		time.Sleep(3 * time.Second)
// 	}
// }

// // Private helper to avoid repeating dial logic
// func (k *KafkaPublisher) dial(ctx context.Context) (*kafka.Conn, error) {
// 	return kafka.DialContext(ctx, "tcp", k.brokers[0])
// }

// func (k *KafkaPublisher) verify(required []string) bool {
// 	conn, err := k.dial(context.Background())
// 	if err != nil {
// 		return false
// 	}
// 	defer conn.Close()

// 	partitions, err := conn.ReadPartitions()
// 	if err != nil {
// 		return false
// 	}

// 	existing := make(map[string]bool)
// 	for _, p := range partitions {
// 		existing[p.Topic] = true
// 	}

// 	for _, r := range required {
// 		if !existing[r] {
// 			return false
// 		}
// 	}
// 	return true
// }
