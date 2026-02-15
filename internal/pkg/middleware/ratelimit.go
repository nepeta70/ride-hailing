package middleware

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryRateLimit(r rate.Limit, b int, metrics telemetry.MetricsInterface) grpc.UnaryServerInterceptor {
	limiter := rate.NewLimiter(r, b)
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if !limiter.Allow() {
			metrics.RateLimitDrop(info.FullMethod)
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}
		return handler(ctx, req)
	}
}
