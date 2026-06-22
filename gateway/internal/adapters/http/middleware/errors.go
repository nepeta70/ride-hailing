package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recovery(telemetry ports.TelemetryProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				ctx, span := telemetry.Tracer().Start(c.Request.Context(), "HTTP.Recovery", trace.WithSpanKind(trace.SpanKindServer))
				defer span.End()

				telemetry.Logger().ErrorContext(ctx, "HTTP panic recovered", "error", recovered, "path", c.FullPath())
				span.SetStatus(otelcodes.Error, "panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

func WriteGRPCError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(grpcCodeToHTTP(st.Code()), gin.H{"error": st.Message()})
}

func grpcCodeToHTTP(code grpcCodes.Code) int {
	switch code {
	case grpcCodes.OK:
		return http.StatusOK
	case grpcCodes.Canceled:
		return http.StatusRequestTimeout
	case grpcCodes.InvalidArgument, grpcCodes.FailedPrecondition, grpcCodes.OutOfRange:
		return http.StatusBadRequest
	case grpcCodes.NotFound:
		return http.StatusNotFound
	case grpcCodes.AlreadyExists, grpcCodes.Aborted:
		return http.StatusConflict
	case grpcCodes.PermissionDenied:
		return http.StatusForbidden
	case grpcCodes.Unauthenticated:
		return http.StatusUnauthorized
	case grpcCodes.ResourceExhausted:
		return http.StatusTooManyRequests
	case grpcCodes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case grpcCodes.Unimplemented:
		return http.StatusNotImplemented
	case grpcCodes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
