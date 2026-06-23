package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

const (
	headerSenderID    = "X-Sender-Id"
	headerSenderType  = "X-Sender-Type"
	headerSenderName  = "X-Sender-Name"
	headerRequestID   = "X-Request-Id"
	headerCountryCode = "X-Country-Code"
	headerAppVersion  = "X-App-Version"
	headerOS          = "X-OS"
	headerNetwork     = "X-Network"
	headerDeviceID    = "X-Device-Id"
	headerRetryCount  = "X-Retry-Count"
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
		requestID := firstHeader(c, headerRequestID, "request-id")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		senderType := firstHeader(c, headerSenderType, "sender-type")
		if senderType != "" && !enums.SenderType(senderType).IsValid() {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid sender type"})
			return
		}

		ctx := &RequestContext{
			SenderID:    firstHeader(c, headerSenderID, "sender-id"),
			SenderType:  senderType,
			SenderName:  firstHeader(c, headerSenderName, "sender-name"),
			RequestID:   requestID,
			CountryCode: firstHeader(c, headerCountryCode, "country-code"),
			AppVersion:  firstHeader(c, headerAppVersion, "app-version"),
			OS:          firstHeader(c, headerOS, "os"),
			Network:     firstHeader(c, headerNetwork, "network"),
			DeviceID:    firstHeader(c, headerDeviceID, "device-id"),
			RetryCount:  firstHeader(c, headerRetryCount, "retry-count"),
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}

		c.Set("request_context", ctx)
		c.Header(headerRequestID, requestID)
		c.Next()
	}
}

func GetRequestContext(c *gin.Context) (*RequestContext, bool) {
	value, ok := c.Get("request_context")
	if !ok {
		return nil, false
	}
	ctx, ok := value.(*RequestContext)
	return ctx, ok
}

func (r *RequestContext) ToMetadata(apiKey string) metadata.MD {
	return metadata.New(map[string]string{
		"api-key":      apiKey,
		"sender-id":    r.SenderID,
		"sender-type":  r.SenderType,
		"sender-name":  r.SenderName,
		"request-id":   r.RequestID,
		"timestamp":    r.Timestamp,
		"country-code": r.CountryCode,
		"app-version":  r.AppVersion,
		"os":           r.OS,
		"network":      r.Network,
		"device-id":    r.DeviceID,
		"retry-count":  r.RetryCount,
	})
}

func OutgoingGRPCContext(c *gin.Context, apiKey string, propagator propagation.TextMapPropagator) context.Context {
	reqCtx, _ := GetRequestContext(c)
	if reqCtx == nil {
		reqCtx = &RequestContext{
			RequestID: uuid.NewString(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	ctx := c.Request.Context()
	if propagator != nil {
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))
	}

	return metadata.NewOutgoingContext(ctx, reqCtx.ToMetadata(apiKey))
}

func firstHeader(c *gin.Context, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(c.GetHeader(name)); value != "" {
			return value
		}
	}
	return ""
}
