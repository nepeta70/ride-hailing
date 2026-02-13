package middleware

import (
	"context"
	"strconv"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ContextInterceptor(cm *ctxmgr.ContextManager, config *config.BaseConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 1. Security Check (Fail Fast)
		apiKey := getMetadata(md, "x-api-key")
		if apiKey != config.APIKey {
			return nil, status.Error(codes.Unauthenticated, "invalid api key")
		}

		userID := getMetadata(md, "user-id")
		if userID == "" {
			return nil, status.Error(codes.Unauthenticated, "missing user ID")
		}

		requestID := getMetadata(md, "x-request-id")
		if requestID == "" {
			return nil, status.Error(codes.InvalidArgument, "missing request ID")
		}

		// 2. Assemble the RequestInfo
		rInfo := ctxmgr.RequestInfo{
			User: ctxmgr.UserSession{
				ID:   userID,
				Role: getMetadata(md, "user-role"), // rider, driver, admin
			},
			Trace: ctxmgr.TraceInfo{
				RequestID:  requestID,
				Origin:     getMetadata(md, "x-origin-service"),
				Timestamp:  getMetadata(md, "x-timestamp"),
				RetryCount: getIntMetadata(md, "x-retry-count"),
			},
			Location: ctxmgr.LocationInfo{
				CountryCode: getMetadata(md, "x-country-code"),
			},
			Client: ctxmgr.ClientInfo{
				AppVersion: getMetadata(md, "x-app-version"),
				OS:         getMetadata(md, "x-os"),
				Network:    getMetadata(md, "x-network"),
				DeviceID:   getMetadata(md, "x-device-id"),
			},
		}

		// 3. Inject and Continue
		return handler(cm.Inject(ctx, rInfo), req)
	}
}

func getMetadata(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func getIntMetadata(md metadata.MD, key string) int {
	if vals := md.Get(key); len(vals) > 0 {
		if n, err := strconv.Atoi(vals[0]); err == nil {
			return n
		}
	}
	return 0
}
