package middleware

import (
	"context"
	"strings"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

func FilteredChain(cfg *config.BaseConfig, logger ports.Logger, contextManager *ctxmgr.ContextManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// 1. Check if we should skip (Health Check)
		if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
			return handler(ctx, req)
		}

		interceptors := []grpc.UnaryServerInterceptor{
			UnaryTimeout(cfg.Server.WriteTimeout),
			UnaryRateLimit(rate.Limit(cfg.Security.RateLimit), cfg.Security.RateBurst),
			ContextInterceptor(contextManager, cfg),
			UnaryServerLogging(logger),
		}
		// 2. Build the execution chain for business logic
		// We nest the interceptors so they execute in the order provided
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextHandler := chain
			chain = func(currentCtx context.Context, currentReq any) (any, error) {
				return currentInterceptor(currentCtx, currentReq, info, nextHandler)
			}
		}

		return chain(ctx, req)
	}
}
