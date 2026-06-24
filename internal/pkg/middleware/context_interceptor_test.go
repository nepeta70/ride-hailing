package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/nepeta70/ride-hailing/internal/pkg/auth"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func signedMetadata(fields map[string]string, secret string) metadata.MD {
	md := metadata.New(fields)
	return auth.AttachSignature(md, secret)
}

func TestContextInterceptor(t *testing.T) {
	cfg := config.DefaultBaseConfig()
	cfg.APIKey = "test-secret"
	cfg.HMACSecret = "test-hmac-secret"
	cm := ctxmgr.NewContextManager()
	telemetryProvider := mocks.NewMockTelemetryProvider()

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	endpointRoles := &mocks.EndpointRequests{}
	contextInterceptor, _ := NewContextInterceptor(&ContextInterceptorOptions{
		ContextManager: cm,
		Config:         &cfg,
		Telemetry:      telemetryProvider,
		EndpointRoles:  endpointRoles,
	})

	validUserID := uuid.New().String()
	validReqID := uuid.New().String()
	validTimestamp := time.Now().UTC().Format(time.RFC3339)

	baseFields := map[string]string{
		"api-key":      "test-secret",
		"sender-id":    validUserID,
		"sender-type":  "rider",
		"request-id":   validReqID,
		"country-code": "ES",
		"timestamp":    validTimestamp,
	}

	tests := []struct {
		name         string
		md           metadata.MD
		expectedCode codes.Code
		expectedRole enums.SenderType
	}{
		{
			name:         "Success - Full valid metadata",
			md:           signedMetadata(baseFields, cfg.HMACSecret),
			expectedCode: codes.OK,
			expectedRole: enums.SenderType("rider"),
		},
		{
			name: "Failure - Missing Timestamp",
			md: signedMetadata(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
			}, cfg.HMACSecret),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Invalid Timestamp (non-numeric)",
			md: signedMetadata(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    "notanumber",
			}, cfg.HMACSecret),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Invalid Timestamp (too old)",
			md: signedMetadata(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			}, cfg.HMACSecret),
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name: "Failure - Invalid Timestamp (in the future)",
			md: signedMetadata(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			}, cfg.HMACSecret),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Invalid API Key",
			md: metadata.New(map[string]string{
				"api-key": "wrong-key",
			}),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Missing User ID",
			md: signedMetadata(map[string]string{
				"api-key":     "test-secret",
				"sender-type": "rider",
				"request-id":  validReqID,
				"timestamp":   validTimestamp,
			}, cfg.HMACSecret),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Invalid User Role",
			md: signedMetadata(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "invalid-role",
				"request-id":  validReqID,
				"timestamp":   validTimestamp,
			}, cfg.HMACSecret),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Missing Request ID",
			md: signedMetadata(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "rider",
				"timestamp":   validTimestamp,
			}, cfg.HMACSecret),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Missing Signature",
			md: metadata.New(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "rider",
				"request-id":  validReqID,
				"timestamp":   validTimestamp,
			}),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Invalid Signature",
			md: metadata.New(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "rider",
				"request-id":  validReqID,
				"timestamp":   validTimestamp,
				"signature":   "invalid-signature",
			}),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Tampered Sender ID",
			md: func() metadata.MD {
				md := signedMetadata(baseFields, cfg.HMACSecret)
				md.Set("sender-id", uuid.New().String())
				return md
			}(),
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), tt.md)

			handler := func(currentCtx context.Context, req any) (any, error) {
				if tt.expectedCode == codes.OK {
					rInfo, ok := cm.Extract(currentCtx)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedRole, rInfo.Sender.Type)
					assert.Equal(t, validUserID, rInfo.Sender.ID.String())
				}
				return "resp", nil
			}

			interceptor := contextInterceptor.Unary()
			_, err := interceptor(ctx, nil, info, handler)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
