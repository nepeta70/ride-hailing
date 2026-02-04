package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryTimeout_Table(t *testing.T) {
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
			mw := UnaryTimeout(tc.timeout)
			resp, err := mw(context.Background(), nil, nil, tc.handler)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantCode != codes.Unknown {
					assert.Equal(t, tc.wantCode, status.Code(err))
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestUnaryRateLimit_Table(t *testing.T) {
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
				mw := middleware.UnaryRateLimit(rate.Every(time.Second), 1)
				handler := func(ctx context.Context, req any) (any, error) {
					return "allowed", nil
				}
				return func(ctx context.Context, req any, _ *struct{}, h func(context.Context, any) (any, error)) (any, error) {
					return mw(ctx, req, nil, h)
				}, handler
			},
			wantResp: "allowed",
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "exceed",
			setup: func() (func(context.Context, any, *struct{}, func(context.Context, any) (any, error)) (any, error), func(context.Context, any) (any, error)) {
				mw := middleware.UnaryRateLimit(rate.Every(time.Second), 1)
				handler := func(ctx context.Context, req any) (any, error) {
					return "should not run", nil
				}
				// First call to fill the token bucket
				_, _ = mw(context.Background(), nil, nil, handler)
				return func(ctx context.Context, req any, _ *struct{}, h func(context.Context, any) (any, error)) (any, error) {
					return mw(ctx, req, nil, h)
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
