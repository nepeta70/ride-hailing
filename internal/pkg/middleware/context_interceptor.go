package middleware

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ContextInterceptorOptions struct {
	ContextManager *ctxmgr.ContextManager
	Config         *config.BaseConfig
	Telemetry      ports.TelemetryProvider
	EndpointRoles  ports.EndpointRoles
}

func (o *ContextInterceptorOptions) Validate() error {
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if o.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
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
	telemetry      ports.TelemetryProvider
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
		telemetry:      options.Telemetry,
		endpointRoles:  options.EndpointRoles,
	}, nil
}

// Unary provides the gRPC interceptor logic for identity and context assembly.
func (i *ContextInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tr := i.telemetry.Tracer()
		ctx, span := tr.Start(ctx, "Middleware.ContextInterceptor", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		i.telemetry.Logger().Debug("gRPC request received", "method", info.FullMethod, "payload", req, "timestamp", time.Now().UTC().Format(time.RFC3339))
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.telemetry.Metrics().AuthFailure(info.FullMethod, "missing_metadata")
			i.telemetry.Logger().Warn("Missing metadata in request", "method", info.FullMethod)
			span.SetStatus(codes.Error, "missing_metadata")
			return nil, errUnauthenticated
		}

		i.telemetry.Logger().Debug("Received metadata:", "metadata", md)

		// 1. Security Check (Fail Fast)
		apiKey := getMetadata(md, "api-key")
		if apiKey != i.config.APIKey {
			i.telemetry.Metrics().AuthFailure(info.FullMethod, "invalid_api_key")
			i.telemetry.Logger().Warn("Invalid API Key", "method", info.FullMethod)
			span.SetAttributes(
				attribute.String("auth.reason", "invalid_api_key"),
				attribute.String("auth.received_key", apiKey),
			)
			return nil, errUnauthenticated
		}

		senderID := getUUIDMetadata(md, "sender-id")
		if senderID == uuid.Nil {
			i.telemetry.Metrics().AuthFailure(info.FullMethod, "missing_user_id")
			i.telemetry.Logger().Warn("Missing sender ID", "method", info.FullMethod)
			span.SetStatus(codes.Error, "missing_user_id")
			return nil, errUnauthenticated
		}

		senderType := getMetadata(md, "sender-type")
		role := enums.SenderType(senderType)
		if !role.IsValid() {
			i.telemetry.Metrics().AuthFailure(info.FullMethod, "invalid_role")
			i.telemetry.Logger().Warn("Invalid sender type", "method", info.FullMethod, "sender_type", senderType)
			span.SetStatus(codes.Error, "invalid_role")
			return nil, errUnauthenticated
		}

		requestID := getUUIDMetadata(md, "request-id")
		if requestID == uuid.Nil {
			i.telemetry.Metrics().ValidationFailure(info.FullMethod, "missing_request_id")
			i.telemetry.Logger().Warn("Missing request ID", "method", info.FullMethod)
			span.SetStatus(codes.Error, "missing_request_id")
			return nil, errInvalidArgument
		}

		// Role validation
		rolesForRequest := i.endpointRoles.RequestRoles()
		if len(rolesForRequest) > 0 {
			if roles, ok := rolesForRequest[info.FullMethod]; ok {
				if !slices.Contains(roles, role) {
					i.telemetry.Logger().Warn("Permission Denied",
						"method", info.FullMethod,
						"required_roles", roles,
						"provided_role", role)
					span.SetAttributes(
						attribute.String("auth.reason", "role_mismatch"),
						attribute.String("auth.role.provided", role.String()),
					)
					i.telemetry.Metrics().AuthFailure(info.FullMethod, "invalid_role")
					return nil, errPermissionDenied
				}
			}
		}

		timestamp, err := i.extractTimestamp(info.FullMethod, md)
		if err != nil {
			span.SetStatus(codes.Error, "invalid_timestamp")
			return nil, err
		}

		// 2. Assemble the RequestInfo (Pointer-based to minimize boxing cost)
		rInfo := &ctxmgr.RequestInfo{
			Sender: ctxmgr.Sender{
				ID:   senderID,
				Type: role,
				Name: getMetadata(md, "sender-name"),
			},
			Trace: ctxmgr.TraceInfo{
				RequestID:  requestID,
				Timestamp:  *timestamp,
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

		span.SetStatus(codes.Ok, "context assembled successfully")
		// 3. Inject and Continue
		return handler(i.contextManager.Inject(ctx, rInfo), req)
	}
}

func (i *ContextInterceptor) extractTimestamp(method string, md metadata.MD) (*time.Time, error) {
	sTimestamp := getMetadata(md, "timestamp")
	if len(sTimestamp) == 0 {
		i.telemetry.Metrics().ValidationFailure(method, "missing_timestamp")
		return nil, errInvalidArgument
	}
	timestamp, err := time.Parse(time.RFC3339, sTimestamp)
	if err != nil {
		i.telemetry.Logger().Error("Invalid timestamp", "error", err, "timestamp", sTimestamp)
		i.telemetry.Metrics().ValidationFailure(method, "invalid_timestamp")
		return nil, errInvalidArgument
	}

	now := time.Now().UTC()
	if timestamp.After(now.Add(i.config.Timeouts.MaxClockDrift)) {
		i.telemetry.Metrics().ValidationFailure(method, "timestamp_too_far_in_future")
		return nil, errInvalidArgument
	}

	if timestamp.Before(now.Add(-i.config.Timeouts.RequestTimeout)) {
		i.telemetry.Metrics().ValidationFailure(method, "timestamp_expired")
		return nil, errDeadlineExceeded
	}
	return &timestamp, nil
}
