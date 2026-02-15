package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func BenchmarkInterceptorChain_Overhead(b *testing.B) {
	// Setup
	opts := &FilteredChainOpts{
		Config: &config.BaseConfig{
			Server:   config.ServerConfig{WriteTimeout: 5 * time.Second},
			Security: config.SecurityConfig{RateLimit: 1000000, RateBurst: 1000000}, // High limit to avoid throttling
		},
		Logger:            &mocks.MockLogger{},
		ContextManager:    ctxmgr.NewContextManager(),
		AuthConfiguration: &mocks.EndpointRequests{}, // No roles required for test
		Metrics:           &mocks.MockMetrics{},      // No-op metrics
	}

	chain, _ := NewInterceptorChain(opts)
	filteredInterceptor := chain.FilteredChain()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/ride.RideService/GetRide",
	}

	ctx := context.Background()

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

func BenchmarkInterceptorChain_OverheadUsingPrometheus(b *testing.B) {
	reg := prometheus.NewRegistry()

	// Setup
	opts := &FilteredChainOpts{
		Config: &config.BaseConfig{
			Server:   config.ServerConfig{WriteTimeout: 5 * time.Second},
			Security: config.SecurityConfig{RateLimit: 1000000, RateBurst: 1000000}, // High limit to avoid throttling
		},
		Logger:            &mocks.MockLogger{},
		ContextManager:    ctxmgr.NewContextManager(),
		AuthConfiguration: &mocks.EndpointRequests{},                                     // No roles required for test
		Metrics:           telemetry.NewMetrics("test_namespace", "test_subsystem", reg), // Use real metrics
	}

	chain, _ := NewInterceptorChain(opts)
	filteredInterceptor := chain.FilteredChain()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/ride.RideService/GetRide",
	}

	ctx := context.Background()

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
