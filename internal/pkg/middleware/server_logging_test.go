package middleware_test

import (
	"context"
	"testing"

	. "github.com/nepeta70/ride-hailing/internal/pkg/middleware"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerLogging_Table(t *testing.T) {
	type args struct {
		handler func(ctx context.Context, req any) (any, error)
		info    *grpc.UnaryServerInfo
	}
	metrics := &mocks.MockMetrics{}
	tests := []struct {
		name      string
		args      args
		wantResp  any
		wantErr   bool
		wantCode  codes.Code
		wantError bool // expect error log
	}{
		{
			name: "normal",
			args: args{
				handler: func(ctx context.Context, req any) (any, error) { return "ok", nil },
				info:    &grpc.UnaryServerInfo{FullMethod: "/test"},
			},
			wantResp:  "ok",
			wantErr:   false,
			wantCode:  codes.OK,
			wantError: false,
		},
		{
			name: "error",
			args: args{
				handler: func(ctx context.Context, req any) (any, error) {
					return nil, status.Error(codes.InvalidArgument, "bad input")
				},
				info: &grpc.UnaryServerInfo{FullMethod: "/err"},
			},
			wantResp:  nil,
			wantErr:   true,
			wantCode:  codes.InvalidArgument,
			wantError: false,
		},
		{
			name: "panic",
			args: args{
				handler: func(ctx context.Context, req any) (any, error) { panic("fail") },
				info:    &grpc.UnaryServerInfo{FullMethod: "/panic"},
			},
			wantResp:  nil,
			wantErr:   true,
			wantCode:  codes.Internal,
			wantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mocks.MockLogger{}
			mw := UnaryServerLogging(logger, metrics)
			resp, err := mw(context.Background(), nil, tc.args.info, tc.args.handler)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantCode, status.Code(err))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantResp, resp)
			if tc.wantError {
				foundError := false
				for _, entry := range logger.Entries {
					if len(entry) >= 6 && entry[:6] == "ERROR:" {
						foundError = true
					}
				}
				assert.True(t, foundError, "expected error log, got %v", logger.Entries)
			}
			// Always expect an info log
			assert.NotEmpty(t, logger.Entries)
			foundInfo := false
			for _, entry := range logger.Entries {
				if len(entry) >= 5 && entry[:5] == "INFO:" {
					foundInfo = true
				}
			}
			assert.True(t, foundInfo, "expected info log, got %v", logger.Entries)
		})
	}
}
