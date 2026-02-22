package middleware

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ContextInterceptorOptions struct {
	ContextManager *ctxmgr.ContextManager
	Config         *config.BaseConfig
	Logger         ports.Logger
	Metrics        ports.Metrics
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
	metrics        ports.Metrics
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

		i.logger.Debug("Received metadata:", "metadata", md)

		// 1. Security Check (Fail Fast)
		apiKey := getMetadata(md, "api-key")
		if apiKey != i.config.APIKey {
			i.metrics.AuthFailure(info.FullMethod, "invalid_api_key")
			return nil, errUnauthenticated
		}

		userID := getUUIDMetadata(md, "sender-id")
		if userID == uuid.Nil {
			i.metrics.AuthFailure(info.FullMethod, "missing_user_id")
			return nil, errUnauthenticated
		}

		userRoleStr := getMetadata(md, "sender-type")
		role := enums.SenderType(userRoleStr)
		if !role.IsValid() {
			i.metrics.AuthFailure(info.FullMethod, "invalid_role")
			return nil, errUnauthenticated
		}

		requestID := getUUIDMetadata(md, "request-id")
		if requestID == uuid.Nil {
			i.metrics.ValidationFailure(info.FullMethod, "missing_request_id")
			return nil, errInvalidArgument
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

		timestamp, err := i.extractTimestamp(info.FullMethod, md)
		if err != nil {
			return nil, err
		}

		// 2. Assemble the RequestInfo (Pointer-based to minimize boxing cost)
		rInfo := &ctxmgr.RequestInfo{
			Sender: ctxmgr.Sender{
				ID:   userID,
				Role: role,
				Name: getMetadata(md, "sender-name"),
			},
			Trace: ctxmgr.TraceInfo{
				RequestID:  requestID,
				Timestamp:  timestamp,
				RetryCount: getIntMetadata(md, "retry-count"),
			},
			Location: ctxmgr.LocationInfo{
				CountryCode: getMetadata(md, "country-code"),
			},
			Client: ctxmgr.ClientInfo{
				AppVersion: getMetadata(md, "app-version"),
				OS:         getMetadata(md, "os"),
				Network:    getMetadata(md, "network"),
				DeviceID:   getMetadata(md, "device-id"),
			},
		}

		// 3. Inject and Continue
		return handler(i.contextManager.Inject(ctx, rInfo), req)
	}
}

func (i *ContextInterceptor) extractTimestamp(method string, md metadata.MD) (int64, error) {
	sTimestamp := getMetadata(md, "timestamp")
	if len(sTimestamp) == 0 {
		i.metrics.ValidationFailure(method, "missing_timestamp")
		return 0, errInvalidArgument
	}
	timestamp, err := strconv.ParseInt(sTimestamp, 10, 64)
	if err != nil {
		i.metrics.ValidationFailure(method, "invalid_timestamp")
		return 0, errInvalidArgument
	}

	if timestamp == 0 {
		i.metrics.ValidationFailure(method, "missing_timestamp")
		return 0, errInvalidArgument
	}

	now := time.Now().Unix()
	if timestamp > (now + int64(i.config.Timeouts.MaxClockDriftSeconds)) {
		i.metrics.ValidationFailure(method, "timestamp_too_far_in_future")
		return 0, errInvalidArgument
	}

	if timestamp < (now - int64(i.config.Timeouts.RequestTimeoutSeconds)) {
		i.metrics.ValidationFailure(method, "timestamp_expired")
		return 0, errDeadlineExceeded
	}
	return timestamp, nil
}
