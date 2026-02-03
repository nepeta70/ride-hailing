package kafka

import (
	"fmt"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type kafkaLogger struct {
	logger ports.Logger
}

func (l *kafkaLogger) Printf(msg string, args ...any) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}
