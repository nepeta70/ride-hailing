package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/gateway/internal/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/auth"

	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const (
	headerAPIKey      = "api-key"
	headerSenderID    = "sender-id"
	headerSenderType  = "sender-type"
	headerSenderName  = "sender-name"
	headerRequestID   = "request-id"
	headerCountryCode = "country-code"
	headerAppVersion  = "app-version"
	headerOS          = "os"
	headerNetwork     = "network"
	headerDeviceID    = "device-id"
	headerRetryCount  = "retry-count"
	headerTimestamp   = "timestamp"
	headerSignature   = "signature"

	requestContext = "request_context"
)

type HTTPMiddlewareOptions struct {
	Config    *config.Config
	Telemetry ports.TelemetryProvider
}

func (o *HTTPMiddlewareOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
	}
	return nil
}

type RequestContext struct {
	SenderID    string
	SenderType  string
	SenderName  string
	RequestID   string
	CountryCode string
	AppVersion  string
	OS          string
	Network     string
	DeviceID    string
	RetryCount  string
	Timestamp   string
}

func RequestContextMiddleware(options *HTTPMiddlewareOptions) gin.HandlerFunc {
	if err := options.Validate(); err != nil {
		panic(err) // TODO : error
	}
	return func(c *gin.Context) {
		tr := options.Telemetry.Tracer()
		_, span := tr.Start(c.Request.Context(), "Middleware.RequestContext", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		method := c.Request.Method + " " + c.FullPath()
		options.Telemetry.Logger().Debug("HTTP request received",
			"method", method,
			"path", c.Request.URL.Path,
			"timestamp", time.Now().UTC().Format(time.RFC3339),
		)

		// 1. API Key check (Fail Fast)
		apiKey := header(c, headerAPIKey)
		if apiKey != options.Config.APIKey {
			options.Telemetry.Metrics().AuthFailure(method, "invalid_api_key")
			options.Telemetry.Logger().Warn("Invalid API Key", "method", method, "remote_ip", c.ClientIP())
			span.SetAttributes(
				attribute.String("auth.reason", "invalid_api_key"),
				attribute.String("auth.received_key", apiKey),
			)
			span.SetStatus(codes.Error, "invalid_api_key")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		// 2. Sender ID validation
		senderID := header(c, headerSenderID)
		if senderID == "" {
			options.Telemetry.Metrics().AuthFailure(method, "missing_user_id")
			options.Telemetry.Logger().Warn("Missing sender ID", "method", method)
			span.SetStatus(codes.Error, "missing_user_id")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		// 3. Sender type / role validation
		senderType := header(c, headerSenderType)
		role := enums.SenderType(senderType)
		if senderType == "" || !role.IsValid() {
			options.Telemetry.Metrics().AuthFailure(method, "invalid_role")
			options.Telemetry.Logger().Warn("Invalid sender type", "method", method, "sender_type", senderType)
			span.SetStatus(codes.Error, "invalid_role")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid sender type"})
			return
		}

		// 4. Request ID validation
		requestID := header(c, headerRequestID)
		if requestID == "" {
			options.Telemetry.Metrics().ValidationFailure(method, "missing_request_id")
			options.Telemetry.Logger().Warn("Missing request ID", "method", method)
			span.SetStatus(codes.Error, "missing_request_id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request-id"})
			return
		}

		// 6. Timestamp validation
		timestampStr := header(c, headerTimestamp)
		timestamp, err := validateUnixTimestamp(timestampStr)
		if err != nil {
			options.Telemetry.Metrics().AuthFailure(method, "invalid_timestamp")
			options.Telemetry.Logger().Warn("Invalid timestamp", "method", method)
			span.SetStatus(codes.Error, "invalid_timestamp")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := header(c, headerSignature)
		canonicalPayload := auth.CanonicalPayload(senderID, senderType, requestID, timestampStr)
		if !auth.Verify(options.Config.HMACSecret, canonicalPayload, signature) {
			options.Telemetry.Metrics().AuthFailure(method, "invalid_signature")
			options.Telemetry.Logger().Warn("Invalid signature", "method", method, "signature", signature)
			span.SetStatus(codes.Error, "invalid_signature")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}

		reqCtx := &RequestContext{
			SenderID:    senderID,
			SenderType:  senderType,
			SenderName:  header(c, headerSenderName),
			RequestID:   requestID,
			CountryCode: header(c, headerCountryCode),
			AppVersion:  header(c, headerAppVersion),
			OS:          header(c, headerOS),
			Network:     header(c, headerNetwork),
			DeviceID:    header(c, headerDeviceID),
			RetryCount:  header(c, headerRetryCount),
			Timestamp:   timestamp.Format(time.RFC3339),
		}

		c.Set(requestContext, reqCtx)
		c.Header(headerRequestID, requestID)
		c.Next()
	}
}

func GetRequestContext(c *gin.Context) (*RequestContext, bool) {
	value, ok := c.Get(requestContext)
	if !ok {
		return nil, false
	}
	ctx, ok := value.(*RequestContext)
	return ctx, ok
}

func (r *RequestContext) ToMetadata(apiKey, hmacSecret string) metadata.MD {
	md := make(metadata.MD, 11)
	md.Set("api-key", apiKey)
	md.Set("sender-id", r.SenderID)
	md.Set("sender-type", r.SenderType)
	md.Set("sender-name", r.SenderName)
	md.Set("request-id", r.RequestID)
	md.Set("timestamp", r.Timestamp)
	md.Set("country-code", r.CountryCode)
	md.Set("app-version", r.AppVersion)
	md.Set("os", r.OS)
	md.Set("network", r.Network)
	md.Set("device-id", r.DeviceID)
	md.Set("retry-count", r.RetryCount)

	return auth.AttachSignature(md, hmacSecret)
}

func OutgoingGRPCContext(c *gin.Context, apiKey, hmacSecret string, propagator propagation.TextMapPropagator) context.Context {
	reqCtx, ok := GetRequestContext(c)
	if !ok || reqCtx == nil {
		reqCtx = &RequestContext{
			RequestID: uuid.NewString(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	ctx := c.Request.Context()
	if propagator != nil {
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
	}

	return metadata.NewOutgoingContext(ctx, reqCtx.ToMetadata(apiKey, hmacSecret))
}

func header(c *gin.Context, name string) string {
	return strings.TrimSpace(c.GetHeader(name))
}

func validateUnixTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0).UTC(), nil
}
