package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func BenchmarkInterceptorChain_Overhead(b *testing.B) {
	reg := prometheus.NewRegistry()

	cfg := &config.BaseConfig{
		Server:   config.ServerConfig{WriteTimeout: 5 * time.Second},
		Security: config.SecurityConfig{RateLimit: 1000000, RateBurst: 1000000}, // High limit to avoid throttling
	}

	//logger := cfg.Logging.ConfigureLogger()
	// Setup
	opts := &FilteredChainOpts{
		Config:         cfg,
		Logger:         &mocks.MockLogger{}, // Assuming this is a no-op logger for testing
		ContextManager: ctxmgr.NewContextManager(),
		EndpointRoles:  &mocks.EndpointRequests{},                                     // No roles required for test
		Metrics:        telemetry.NewMetrics("test_namespace", "test_subsystem", reg), // Use real metrics
	}

	chain, _ := NewInterceptorChain(opts)
	filteredInterceptor := chain.FilteredChain()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/ride.RideService/GetRide",
	}

	md := metadata.New(map[string]string{
		"x-api-key":        "test-secret-key",
		"sender-id":        uuid.New().String(),
		"sender-type":      "driver",
		"x-request-id":     uuid.New().String(),
		"x-origin-service": "bench-tool",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Benchmark the "Naked" Handler (Baseline)
	b.Run("Baseline_No_Interceptors", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = handler(ctx, "req")
		}
	})

	// Benchmark the Full Interceptor Chain
	b.Run("Full_Chain_Overhead", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = filteredInterceptor(ctx, "req", info, handler)
		}
	})
}
