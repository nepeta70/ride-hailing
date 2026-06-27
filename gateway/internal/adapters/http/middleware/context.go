package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/auth"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

const (
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

	requestContext = "request_context"
)

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

func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := header(c, headerRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		senderType := header(c, headerSenderType)
		if senderType != "" && !enums.SenderType(senderType).IsValid() {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid sender type"})
			return
		}

		ctx := &RequestContext{
			SenderID:    header(c, headerSenderID),
			SenderType:  senderType,
			SenderName:  header(c, headerSenderName),
			RequestID:   requestID,
			CountryCode: header(c, headerCountryCode),
			AppVersion:  header(c, headerAppVersion),
			OS:          header(c, headerOS),
			Network:     header(c, headerNetwork),
			DeviceID:    header(c, headerDeviceID),
			RetryCount:  header(c, headerRetryCount),
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}

		c.Set(requestContext, ctx)
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
