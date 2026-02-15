package middleware

import (
	"context"
	"strings"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type FilteredChainOpts struct {
	Config                 *config.BaseConfig
	Logger                 ports.Logger
	ContextManager         *ctxmgr.ContextManager
	AuthConfiguration      ports.EndpointRoles
	AdditionalInterceptors []grpc.UnaryServerInterceptor
	Metrics                telemetry.MetricsInterface
}

func (opts *FilteredChainOpts) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if opts.AuthConfiguration == nil {
		return errors.NewValidationErrorf("auth configuration is required")
	}
	if opts.Metrics == nil {
		return errors.NewValidationErrorf("metrics is required")
	}
	return nil
}

type InterceptorChain struct {
	interceptors     []grpc.UnaryServerInterceptor
	opts             *FilteredChainOpts
	finalInterceptor grpc.UnaryServerInterceptor
}

func NewInterceptorChain(opts *FilteredChainOpts) (*InterceptorChain, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	cb, err := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())
	if err != nil {
		return nil, err
	}

	interceptors := []grpc.UnaryServerInterceptor{
		UnaryServerLogging(opts.Logger, opts.Metrics),
		ContextInterceptor(opts.ContextManager, opts.Config, opts.Logger, opts.Metrics),
		RoleBasedAccessInterceptor(opts.AuthConfiguration, opts.ContextManager, opts.Metrics),
		UnaryTimeout(opts.Config.Server.WriteTimeout, opts.Metrics),
		UnaryRateLimit(rate.Limit(opts.Config.Security.RateLimit), opts.Config.Security.RateBurst, opts.Metrics),
		UnaryCircuitBreaker(cb, opts.Metrics),
	}
	if len(opts.AdditionalInterceptors) > 0 {
		interceptors = append(interceptors, opts.AdditionalInterceptors...)
	}

	return &InterceptorChain{
		interceptors:     interceptors,
		opts:             opts,
		finalInterceptor: chainInterceptors(interceptors),
	}, nil
}

func (ic *InterceptorChain) FilteredChain() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// 1. Check if we should skip (Health Check)
		if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") {
			res, err := handler(ctx, req)
			if err != nil {
				ic.opts.Logger.Error("Health check failed", "error", err)
			}
			return res, err
		}

		return ic.finalInterceptor(ctx, req, info, handler)
	}
}

func chainInterceptors(interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		buildChain := func(currentI grpc.UnaryServerInterceptor, nextH grpc.UnaryHandler) grpc.UnaryHandler {
			return func(c context.Context, r any) (any, error) {
				return currentI(c, r, info, nextH)
			}
		}

		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildChain(interceptors[i], chain)
		}
		return chain(ctx, req)
	}
}
