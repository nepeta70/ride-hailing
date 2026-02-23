package middleware

import (
	"context"
	"strings"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
)

type FilteredChainOpts struct {
	Config                 *config.BaseConfig
	ContextManager         *ctxmgr.ContextManager
	EndpointRoles          ports.EndpointRoles
	AdditionalInterceptors []grpc.UnaryServerInterceptor
	Telemetry              ports.TelemetryProvider
}

func (opts *FilteredChainOpts) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if opts.EndpointRoles == nil {
		return errors.NewValidationErrorf("auth configuration is required")
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

	recoveryInterceptor := NewRecoveryInterceptor(opts.Telemetry.Logger())
	contextInterceptor, err := NewContextInterceptor(&ContextInterceptorOptions{
		ContextManager: opts.ContextManager,
		Config:         opts.Config,
		Telemetry:      opts.Telemetry,
		EndpointRoles:  opts.EndpointRoles,
	})
	if err != nil {
		return nil, err
	}
	timeoutInterceptor := NewTimeoutInterceptor(opts.Config.Server.WriteTimeout, opts.Telemetry.Metrics())
	resiliencyInterceptor, err := NewResiliencyInterceptor(opts.Config.Security.RateLimit, opts.Config.Security.RateBurst, opts.Telemetry.Metrics())
	if err != nil {
		return nil, err
	}
	interceptors := []grpc.UnaryServerInterceptor{
		recoveryInterceptor.Unary(),
		contextInterceptor.Unary(),
		timeoutInterceptor.Unary(),
		resiliencyInterceptor.Unary(),
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
				ic.opts.Telemetry.Logger().Error("Health check failed", "error", err)
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
