package middleware

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ContextInterceptorOptions struct {
	ContextManager *ctxmgr.ContextManager
	Config         *config.BaseConfig
	Logger         ports.Logger
	Metrics        telemetry.MetricsInterface
	EndpointRoles  ports.EndpointRoles
}

func (o *ContextInterceptorOptions) Validate() error {
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if o.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if o.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	if o.Metrics == nil {
		return errors.NewValidationErrorf("metrics is required")
	}
	if o.EndpointRoles == nil {
		return errors.NewValidationErrorf("endpoint roles configuration is required")
	}
	return nil
}

// ContextInterceptor handles identity extraction, security checks, and role validation.
type ContextInterceptor struct {
	contextManager *ctxmgr.ContextManager
	config         *config.BaseConfig
	logger         ports.Logger
	metrics        telemetry.MetricsInterface
	endpointRoles  ports.EndpointRoles
}

// NewContextInterceptor initializes the interceptor struct with dependencies.
func NewContextInterceptor(options *ContextInterceptorOptions) (*ContextInterceptor, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	return &ContextInterceptor{
		contextManager: options.ContextManager,
		config:         options.Config,
		logger:         options.Logger,
		metrics:        options.Metrics,
		endpointRoles:  options.EndpointRoles,
	}, nil
}

// Unary provides the gRPC interceptor logic for identity and context assembly.
func (i *ContextInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.metrics.AuthFailure(info.FullMethod, "missing_metadata")
			return nil, errUnauthenticated
		}

		// Note: Keeping this for now as per your request,
		// but remember logger variadics are a primary source of allocs.
		i.logger.Info("Received metadata:", "metadata", md)

		// 1. Security Check (Fail Fast)
		apiKey := getMetadata(md, "x-api-key")
		if apiKey != i.config.APIKey {
			i.metrics.AuthFailure(info.FullMethod, "invalid_api_key")
			return nil, errUnauthenticated
		}

		userID := getUUIDMetadata(md, "user-id")
		if userID == uuid.Nil {
			i.metrics.AuthFailure(info.FullMethod, "missing_user_id")
			return nil, errUnauthenticated
		}

		userRoleStr := getMetadata(md, "user-role")
		role := enums.UserRole(userRoleStr)
		if !role.IsValid() {
			i.metrics.AuthFailure(info.FullMethod, "invalid_role")
			return nil, errUnauthenticated
		}

		requestID := getUUIDMetadata(md, "x-request-id")
		if requestID == uuid.Nil {
			i.metrics.AuthFailure(info.FullMethod, "missing_request_id")
			return nil, errUnauthenticated
		}

		// Role validation
		rolesForRequest := i.endpointRoles.RequestRoles()
		if len(rolesForRequest) > 0 {
			if roles, ok := rolesForRequest[info.FullMethod]; ok {
				if !slices.Contains(roles, role) {
					i.metrics.AuthFailure(info.FullMethod, "invalid_role")
					return nil, errPermissionDenied
				}
			}
		}

		// 2. Assemble the RequestInfo (Pointer-based to minimize boxing cost)
		rInfo := &ctxmgr.RequestInfo{
			User: ctxmgr.UserSession{
				ID:   userID,
				Role: role,
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
		return handler(i.contextManager.Inject(ctx, rInfo), req)
	}
}
