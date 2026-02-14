package middleware

import (
	"context"
	"slices"

	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func APIKeyInterceptor(endpointRoles ports.EndpointRoles, ctxMgr *ctxmgr.ContextManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		contextInfo, ok := ctxMgr.Extract(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to extract context info")
		}

		rolesForRequest := endpointRoles.RequestRoles()
		if len(rolesForRequest) > 0 {
			roles, ok := rolesForRequest[info.FullMethod]
			if ok && !slices.Contains(roles, contextInfo.User.Role) {
				return nil, status.Error(codes.PermissionDenied, "user does not have permission to access this endpoint")
			}
		}

		// Key is valid, proceed to the handler
		return handler(ctx, req)
	}
}
