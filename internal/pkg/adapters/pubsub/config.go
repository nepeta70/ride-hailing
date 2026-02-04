package pubsub

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type KafkaConfig struct {
	Brokers       []string `json:"brokers" env:"KAFKA_BROKERS"`
	AutoCreate    bool     `json:"auto_create_topics" env:"KAFKA_AUTO_CREATE_TOPICS"`
	Topics        []string `json:"topics" env:"KAFKA_TOPICS"`
	BatchSize     int      `json:"batch_size" env:"KAFKA_BATCH_SIZE"`
	BatchTimeout  int      `json:"batch_timeout_ms" env:"KAFKA_BATCH_TIMEOUT_MS"`
	EnableLogging bool     `json:"enable_logging" env:"KAFKA_ENABLE_LOGGING"`
	TopicPrefix   string   `json:"topic_prefix" env:"KAFKA_TOPIC_PREFIX"`
}

func DefaultKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:       []string{"localhost:9092"},
		AutoCreate:    true,
		Topics:        []string{"user", "location", "ride", "driver", "matching", "notification", "rider"},
		BatchSize:     100,
		BatchTimeout:  500,
		EnableLogging: false,
		TopicPrefix:   "",
	}
}

func (c *KafkaConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.NewValidationErrorf("at least one Kafka broker must be specified")
	}
	// if len(c.Topics) == 0 {
	// 	return errors.NewValidationErrorf("at least one Kafka topic must be specified")
	// }
	if c.BatchSize <= 0 {
		return errors.NewValidationErrorf("batch size must be greater than zero")
	}

	if c.BatchTimeout < 0 {
		return errors.NewValidationErrorf("batch timeout must be non-negative")
	}
	return nil
}

func (c *KafkaConfig) GetPrefixedTopic(topic string) string {
	if c.TopicPrefix != "" {
		return c.TopicPrefix + "." + topic
	}

	return topic
}
