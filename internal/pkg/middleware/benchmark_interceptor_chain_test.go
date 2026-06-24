package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/auth"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func BenchmarkInterceptorChain_Overhead(b *testing.B) {
	cfg := &config.BaseConfig{
		Server:     config.ServerConfig{WriteTimeout: 5 * time.Second},
		Security:   config.SecurityConfig{RateLimit: 1000000, RateBurst: 1000000},
		APIKey:     "test-secret-key",
		HMACSecret: "test-hmac-secret",
	}

	//logger := cfg.Logging.ConfigureLogger()
	// Setup
	opts := &FilteredChainOpts{
		Config:         cfg,
		ContextManager: ctxmgr.NewContextManager(),
		EndpointRoles:  &mocks.EndpointRequests{},
		Telemetry:      mocks.NewMockTelemetryProvider(),
	}

	chain, _ := NewInterceptorChain(opts)
	filteredInterceptor := chain.FilteredChain()

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/ride.RideService/GetRide",
	}

	md := auth.AttachSignature(metadata.New(map[string]string{
		"api-key":        "test-secret-key",
		"sender-id":      uuid.New().String(),
		"sender-type":    "driver",
		"request-id":     uuid.New().String(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"origin-service": "bench-tool",
	}), cfg.HMACSecret)
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
