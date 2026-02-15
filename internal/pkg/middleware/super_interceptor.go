package middleware

// import (
// 	"context"

// 	"google.golang.org/grpc"
// )

// type GodInterceptor struct {
// 	opts *FilteredChainOpts
// }

// func NewGodInterceptor(opts *FilteredChainOpts) (*GodInterceptor, error) {
// 	if err := opts.Validate(); err != nil {
// 		return nil, err
// 	}
// 	return &GodInterceptor{opts: opts}, nil
// }
// func (i *GodInterceptor) Unary() grpc.UnaryServerInterceptor {
// 	return func(
// 		ctx context.Context,
// 		req any,
// 		info *grpc.UnaryServerInfo,
// 		handler grpc.UnaryHandler,
// 	) (any, error) {
// 		return handler(ctx, req)
// 	}
// }

// func (i *GodInterceptor) ServerLogging() grpc.StreamServerInterceptor {
// 		// UnaryServerLogging(opts.Logger),
// 		// UnaryMetrics(opts.Metrics),
// 		// ContextInterceptor(opts.ContextManager, opts.Config, opts.Logger, opts.Metrics),
// 		// RoleBasedAccessInterceptor(opts.AuthConfiguration, opts.ContextManager, opts.Metrics),
// 		// UnaryTimeout(opts.Config.Server.WriteTimeout, opts.Metrics),
// 		// UnaryRateLimit(rate.Limit(opts.Config.Security.RateLimit), opts.Config.Security.RateBurst, opts.Metrics),
// 		// UnaryCircuitBreaker(cb, opts.Metrics),