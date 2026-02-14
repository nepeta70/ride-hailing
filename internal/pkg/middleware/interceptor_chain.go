package middleware

import (
	"context"
	"strings"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type FilteredChainOpts struct {
	Config                 *config.BaseConfig
	Logger                 ports.Logger
	ContextManager         *ctxmgr.ContextManager
	AuthConfiguration      ports.EndpointRoles
	AdditionalInterceptors []grpc.UnaryServerInterceptor
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
	return nil
}

type InterceptorChain struct {
	interceptors []grpc.UnaryServerInterceptor
	opts         *FilteredChainOpts
}

func NewInterceptorChain(opts *FilteredChainOpts) (*InterceptorChain, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	interceptors := []grpc.UnaryServerInterceptor{
		UnaryServerLogging(opts.Logger),
		ContextInterceptor(opts.ContextManager, opts.Config, opts.Logger),
		RoleBasedAccessInterceptor(opts.AuthConfiguration, opts.ContextManager),
		UnaryTimeout(opts.Config.Server.WriteTimeout),
		UnaryRateLimit(rate.Limit(opts.Config.Security.RateLimit), opts.Config.Security.RateBurst),
	}
	if len(opts.AdditionalInterceptors) > 0 {
		interceptors = append(interceptors, opts.AdditionalInterceptors...)
	}

	return &InterceptorChain{
		interceptors: interceptors,
		opts:         opts,
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

		// 2. Build the execution chain for business logic
		// We nest the interceptors so they execute in the order provided
		chain := handler
		for i := len(ic.interceptors) - 1; i >= 0; i-- {
			currentInterceptor := ic.interceptors[i]
			nextHandler := chain
			chain = func(currentCtx context.Context, currentReq any) (any, error) {
				return currentInterceptor(currentCtx, currentReq, info, nextHandler)
			}
		}

		return chain(ctx, req)
	}
}
