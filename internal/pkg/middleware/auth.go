package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func APIKeyInterceptor(validKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		keys := md.Get("x-api-key")
		if len(keys) == 0 || keys[0] != validKey {
			return nil, status.Error(codes.Unauthenticated, "invalid or missing API key")
		}

		// Key is valid, proceed to the handler
		return handler(ctx, req)
	}
}
