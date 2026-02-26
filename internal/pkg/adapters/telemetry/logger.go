package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type LoggerFactory struct {
	config *config.LoggingConfig
}

func NewLoggerFactory(config *config.LoggingConfig) *LoggerFactory {
	return &LoggerFactory{config: config}
}

func (c *LoggerFactory) GetConsoleLogger() slog.Handler {
	opts := &slog.HandlerOptions{
		Level: c.slogLevel(),
	}

	return slog.NewJSONHandler(os.Stdout, opts)
}

func (c *LoggerFactory) CreateLogger(telemetryHandler slog.Handler) ports.Logger {
	multi := slog.NewMultiHandler(c.GetConsoleLogger(), telemetryHandler)
	logger := slog.New(&PreProcessHandler{next: multi})
	slog.SetDefault(logger)
	return logger
}

func (c *LoggerFactory) slogLevel() slog.Level {
	switch strings.ToUpper(c.config.Level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

var funcCache sync.Map

// PreProcessHandler prepares the record for ALL downstream handlers (Console + OTel)
type PreProcessHandler struct {
	next slog.Handler
}

func (h *PreProcessHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *PreProcessHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. Calculate Function Name ONCE for all handlers
	if r.PC != 0 {
		if fn, ok := funcCache.Load(r.PC); ok {
			r.AddAttrs(slog.String("func", fn.(string)))
		} else {
			f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
			name := filepath.Base(f.Function)
			funcCache.Store(r.PC, name)
			r.AddAttrs(slog.String("func", name))
		}
	}

	// 2. Convert Stringer types ONCE for all handlers
	// We build a new record to avoid mutating the original in a MultiHandler setup
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindAny {
			if s, ok := a.Value.Any().(fmt.Stringer); ok {
				a.Value = slog.StringValue(s.String())
			}
		}
		newRecord.AddAttrs(a)
		return true
	})

	return h.next.Handle(ctx, newRecord)
}

func (h *PreProcessHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PreProcessHandler{next: h.next.WithAttrs(attrs)}
}

func (h *PreProcessHandler) WithGroup(name string) slog.Handler {
	return &PreProcessHandler{next: h.next.WithGroup(name)}
}
