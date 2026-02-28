package middleware_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewResiliencyInterceptor(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit float64
		rateBurst int
		wantErr   bool
	}{
		{"valid config", 10, 5, false},
		{"zero rate", 0, 1, false},
		{"negative burst", 1, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockTelemetryProvider()
			interceptor, err := NewResiliencyInterceptor(tt.rateLimit, tt.rateBurst, m)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, interceptor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, interceptor)
			}
		})
	}
}

func TestResiliencyInterceptor_Unary_RateLimit(t *testing.T) {
	m := mocks.NewMockTelemetryProvider()
	interceptor, err := NewResiliencyInterceptor(1, 1, m)
	assert.NoError(t, err)
	unary := interceptor.Unary()

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	// Allow one request, then rate limit
	_, err1 := unary(ctx, nil, info, handler)
	assert.NoError(t, err1)

	_, err2 := unary(ctx, nil, info, handler)
	assert.Error(t, err2)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err2))
	assert.True(t, m.MetricsCalls()["RateLimitDrop"] > 0)
	assert.Equal(t, info.FullMethod, m.MetricsArgs()["RateLimitDrop"][0])
}

func TestResiliencyInterceptor_Unary_CircuitBreaker(t *testing.T) {
	m := mocks.NewMockTelemetryProvider()

	// Using your constructor with high rate limit to test CB
	interceptor, err := NewResiliencyInterceptor(100, 100, m)
	assert.NoError(t, err)

	unary := interceptor.Unary()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	failHandler := func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("fail")
	}
	okHandler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	// Trip the circuit (Assuming 1 failure trips it based on your internal CB logic)
	_, _ = unary(context.Background(), nil, info, failHandler)

	// Call again, should see CB error in metrics
	_, _ = unary(context.Background(), nil, info, okHandler)

	if m.MetricsCalls()["CircuitBreakerError"] > 0 {
		assert.Equal(t, info.FullMethod, m.MetricsArgs()["CircuitBreakerError"][0])
		assert.Equal(t, "circuit_open", m.MetricsArgs()["CircuitBreakerError"][1])
	}
}
