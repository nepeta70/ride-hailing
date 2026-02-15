package middleware_test

import (
	"context"
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryRateLimit_Table(t *testing.T) {
	metrics := &mocks.MockMetrics{}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	tests := []struct {
		name     string
		setup    func() (mw func(context.Context, any, *struct{}, func(context.Context, any) (any, error)) (any, error), handler func(context.Context, any) (any, error))
		wantResp any
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "allow",
			setup: func() (func(context.Context, any, *struct{}, func(context.Context, any) (any, error)) (any, error), func(context.Context, any) (any, error)) {
				mw := UnaryRateLimit(rate.Every(time.Second), 1, metrics)
				handler := func(ctx context.Context, req any) (any, error) {
					return "allowed", nil
				}
				return func(ctx context.Context, req any, _ *struct{}, h func(context.Context, any) (any, error)) (any, error) {
					return mw(ctx, req, info, h)
				}, handler
			},
			wantResp: "allowed",
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "exceed",
			setup: func() (func(context.Context, any, *struct{}, func(context.Context, any) (any, error)) (any, error), func(context.Context, any) (any, error)) {
				mw := UnaryRateLimit(rate.Every(time.Second), 1, metrics)
				handler := func(ctx context.Context, req any) (any, error) {
					return "should not run", nil
				}
				// First call to fill the token bucket
				_, _ = mw(context.Background(), nil, info, handler)
				return func(ctx context.Context, req any, _ *struct{}, h func(context.Context, any) (any, error)) (any, error) {
					return mw(ctx, req, info, h)
				}, handler
			},
			wantResp: nil,
			wantErr:  true,
			wantCode: codes.ResourceExhausted,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw, handler := tc.setup()
			resp, err := mw(context.Background(), nil, nil, handler)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantCode, status.Code(err))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
