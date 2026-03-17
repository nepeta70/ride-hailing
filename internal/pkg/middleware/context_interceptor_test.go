package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

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

func TestContextInterceptor(t *testing.T) {
	// Setup dependencies
	cfg := config.DefaultBaseConfig()
	cfg.APIKey = "test-secret"
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

	tests := []struct {
		name         string
		md           metadata.MD
		expectedCode codes.Code
		expectedRole enums.SenderType
	}{
		{
			name: "Success - Full valid metadata",
			md: metadata.New(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			}),
			expectedCode: codes.OK,
			expectedRole: enums.SenderType("rider"),
		},
		{
			name: "Failure - Missing Timestamp",
			md: metadata.New(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
			}),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Invalid Timestamp (non-numeric)",
			md: metadata.New(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    "notanumber",
			}),
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Invalid Timestamp (too old)",
			md: metadata.New(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			}),
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name: "Failure - Invalid Timestamp (in the future)",
			md: metadata.New(map[string]string{
				"api-key":      "test-secret",
				"sender-id":    validUserID,
				"sender-type":  "rider",
				"request-id":   validReqID,
				"country-code": "ES",
				"timestamp":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			}),
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
			md: metadata.New(map[string]string{
				"api-key":     "test-secret",
				"sender-type": "rider",
				"request-id":  validReqID,
			}),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Invalid User Role",
			md: metadata.New(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "invalid-role",
				"request-id":  validReqID,
			}),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Failure - Missing Request ID",
			md: metadata.New(map[string]string{
				"api-key":     "test-secret",
				"sender-id":   validUserID,
				"sender-type": "rider",
			}),
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with metadata
			ctx := metadata.NewIncomingContext(context.Background(), tt.md)

			// Mock gRPC handler
			handler := func(currentCtx context.Context, req any) (any, error) {
				// For success cases, verify the context was actually injected
				if tt.expectedCode == codes.OK {
					rInfo, ok := cm.Extract(currentCtx)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedRole, rInfo.Sender.Type)
					assert.Equal(t, validUserID, rInfo.Sender.ID.String())
				}
				return "resp", nil
			}

			// Execute Interceptor
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
