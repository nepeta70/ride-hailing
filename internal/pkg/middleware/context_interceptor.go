package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ContextInterceptor(cm *ctxmgr.ContextManager, config *config.BaseConfig, logger ports.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		logger.Info("Received metadata:", "metadata", md)

		// 1. Security Check (Fail Fast)
		apiKey := getMetadata(md, "x-api-key")
		if apiKey != config.APIKey {
			return nil, status.Error(codes.Unauthenticated, "invalid api key")
		}

		userID := getUUIDMetadata(md, "user-id")
		if userID == uuid.Nil {
			return nil, status.Error(codes.Unauthenticated, "missing user ID")
		}

		userRole := getMetadata(md, "user-role")
		if !enums.UserRole(userRole).IsValid() {
			return nil, status.Error(codes.Unauthenticated, "invalid user role")
		}

		requestID := getUUIDMetadata(md, "x-request-id")
		if requestID == uuid.Nil {
			return nil, status.Error(codes.Unauthenticated, "missing request ID")
		}

		// 2. Assemble the RequestInfo
		rInfo := ctxmgr.RequestInfo{
			User: ctxmgr.UserSession{
				ID:   userID,
				Role: enums.UserRole(userRole),
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
