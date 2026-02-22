package ctxmgr_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	. "github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
)

func TestContextManager_Extract(t *testing.T) {
	cm := NewContextManager()

	requestID := uuid.New()
	senderID := uuid.New()

	validRequestInfo := &RequestInfo{
		Sender: Sender{
			ID:   senderID,
			Role: enums.SenderTypeDriver,
			Name: "test-driver",
		},
		Trace: TraceInfo{
			RequestID:  requestID,
			Timestamp:  1234567890,
			RetryCount: 0,
		},
		Location: LocationInfo{
			CountryCode: "US",
		},
		Client: ClientInfo{
			AppVersion: "1.0.0",
			OS:         "iOS",
			Network:    "wifi",
			DeviceID:   "device-123",
		},
	}

	tests := []struct {
		name     string
		ctx      context.Context
		wantInfo *RequestInfo
		wantOk   bool
	}{
		{
			name:     "No RequestInfo in context",
			ctx:      context.Background(),
			wantInfo: nil,
			wantOk:   false,
		},
		{
			name:     "RequestInfo present in context",
			ctx:      cm.Inject(context.Background(), validRequestInfo),
			wantInfo: validRequestInfo,
			wantOk:   true,
		},
		{
			name:     "Extract from context with nil RequestInfo",
			ctx:      cm.Inject(context.Background(), nil),
			wantInfo: nil,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := cm.Extract(tt.ctx)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk && tt.wantInfo != nil {
				assert.NotNil(t, info)
				assert.Equal(t, tt.wantInfo.Sender.ID, info.Sender.ID)
				assert.Equal(t, tt.wantInfo.Sender.Role, info.Sender.Role)
				assert.Equal(t, tt.wantInfo.Sender.Name, info.Sender.Name)
				assert.Equal(t, tt.wantInfo.Trace.RequestID, info.Trace.RequestID)
				assert.Equal(t, tt.wantInfo.Trace.Timestamp, info.Trace.Timestamp)
				assert.Equal(t, tt.wantInfo.Trace.RetryCount, info.Trace.RetryCount)
				assert.Equal(t, tt.wantInfo.Location.CountryCode, info.Location.CountryCode)
				assert.Equal(t, tt.wantInfo.Client.AppVersion, info.Client.AppVersion)
				assert.Equal(t, tt.wantInfo.Client.OS, info.Client.OS)
				assert.Equal(t, tt.wantInfo.Client.Network, info.Client.Network)
				assert.Equal(t, tt.wantInfo.Client.DeviceID, info.Client.DeviceID)
			} else if tt.wantInfo == nil {
				assert.Nil(t, info)
			}
		})
	}
}

func TestContextManager_Inject(t *testing.T) {
	cm := NewContextManager()

	requestID := uuid.New()
	senderID := uuid.New()

	validRequestInfo := &RequestInfo{
		Sender: Sender{
			ID:   senderID,
			Role: enums.SenderTypeRider,
			Name: "test-rider",
		},
		Trace: TraceInfo{
			RequestID:  requestID,
			Timestamp:  9876543210,
			RetryCount: 2,
		},
		Location: LocationInfo{
			CountryCode: "CA",
		},
		Client: ClientInfo{
			AppVersion: "2.0.0",
			OS:         "Android",
			Network:    "4G",
			DeviceID:   "device-456",
		},
	}

	tests := []struct {
		name string
		info *RequestInfo
	}{
		{
			name: "Inject nil RequestInfo",
			info: nil,
		},
		{
			name: "Inject valid RequestInfo",
			info: validRequestInfo,
		},
		{
			name: "Inject empty RequestInfo",
			info: &RequestInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cm.Inject(context.Background(), tt.info)
			assert.NotNil(t, ctx)

			val, ok := cm.Extract(ctx)
			assert.True(t, ok)

			if tt.info == nil {
				assert.Nil(t, val)
			} else {
				assert.NotNil(t, val)
				assert.Equal(t, tt.info.Sender.ID, val.Sender.ID)
				assert.Equal(t, tt.info.Sender.Role, val.Sender.Role)
				assert.Equal(t, tt.info.Sender.Name, val.Sender.Name)
				assert.Equal(t, tt.info.Trace.RequestID, val.Trace.RequestID)
				assert.Equal(t, tt.info.Trace.Timestamp, val.Trace.Timestamp)
				assert.Equal(t, tt.info.Trace.RetryCount, val.Trace.RetryCount)
				assert.Equal(t, tt.info.Location.CountryCode, val.Location.CountryCode)
				assert.Equal(t, tt.info.Client.AppVersion, val.Client.AppVersion)
				assert.Equal(t, tt.info.Client.OS, val.Client.OS)
				assert.Equal(t, tt.info.Client.Network, val.Client.Network)
				assert.Equal(t, tt.info.Client.DeviceID, val.Client.DeviceID)
			}
		})
	}
}
