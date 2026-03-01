package mocks

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type MockTelemetryProvider struct {
	metrics    *MockMetrics
	logger     *MockLogger
	propagator *MockPropagator
	tracer     *MockTracer
}

func NewMockTelemetryProvider() *MockTelemetryProvider {
	return &MockTelemetryProvider{
		metrics:    NewMockMetrics(),
		logger:     NewMockLogger(),
		propagator: &MockPropagator{},
		tracer:     &MockTracer{},
	}
}

func (m *MockTelemetryProvider) Metrics() ports.Metrics {
	return m.metrics
}

func (m *MockTelemetryProvider) Logger() ports.Logger {
	return m.logger
}

func (m *MockTelemetryProvider) Tracer() trace.Tracer {
	return m.tracer
}

func (m *MockTelemetryProvider) Propagator() propagation.TextMapPropagator {
	return m.propagator
}

func (m *MockTelemetryProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockTelemetryProvider) LogEntries() []string {
	return m.logger.Entries
}

func (m *MockTelemetryProvider) MetricsCalls() map[string]int {
	return m.metrics.Calls
}

func (m *MockTelemetryProvider) MetricsArgs() map[string][]any {
	return m.metrics.Args
}

var _ ports.TelemetryProvider = (*MockTelemetryProvider)(nil)

type MockPropagator struct{}

func (m *MockPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
}

func (m *MockPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return ctx
}

func (m *MockPropagator) Fields() []string {
	return []string{}
}

var _ propagation.TextMapPropagator = (*MockPropagator)(nil)

type MockTracer struct {
	embedded.Tracer
	FinishedSpans []*MockSpan
}

func (m *MockTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	s := &MockSpan{name: spanName, Span: trace.SpanFromContext(ctx)}
	return ctx, s
}

var _ trace.Tracer = (*MockTracer)(nil)

type MockSpan struct {
	trace.Span
	name  string
	ended bool
	err   error
}

func (s *MockSpan) End(options ...trace.SpanEndOption)               { s.ended = true }
func (s *MockSpan) RecordError(err error, opts ...trace.EventOption) { s.err = err }
func (s *MockSpan) IsRecording() bool                                { return true }
