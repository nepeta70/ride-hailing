package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryTimeout_Table(t *testing.T) {
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	tests := []struct {
		name     string
		timeout  time.Duration
		handler  func(ctx context.Context, req any) (any, error)
		wantResp any
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name:    "timeout exceeded",
			timeout: 10 * time.Millisecond,
			handler: func(ctx context.Context, req any) (any, error) {
				time.Sleep(50 * time.Millisecond)
				return "late", nil
			},
			wantResp: nil,
			wantErr:  true,
			wantCode: codes.DeadlineExceeded,
		},
		{
			name:    "no timeout",
			timeout: 100 * time.Millisecond,
			handler: func(ctx context.Context, req any) (any, error) {
				return "ok", nil
			},
			wantResp: "ok",
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name:    "handler error",
			timeout: 100 * time.Millisecond,
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, errors.New("handler error")
			},
			wantResp: nil,
			wantErr:  true,
			wantCode: codes.Unknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize metrics with maps to avoid nil assignment panics
			metrics := mocks.NewMockTelemetryProvider()

			mw := NewTimeoutInterceptor(tc.timeout, metrics).Unary()
			resp, err := mw(context.Background(), nil, info, tc.handler)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantCode != codes.Unknown {
					assert.Equal(t, tc.wantCode, status.Code(err))
				}

				// Verify metric was called on timeout
				if tc.wantCode == codes.DeadlineExceeded {
					assert.True(t, metrics.MetricsCalls()["RequestTimeout"] > 0)
					assert.Equal(t, info.FullMethod, metrics.MetricsArgs()["RequestTimeout"][0])
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
