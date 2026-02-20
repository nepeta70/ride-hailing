package ctxmgr

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

type RequestInfo struct {
	Sender   Sender
	Trace    TraceInfo
	Location LocationInfo
	Client   ClientInfo
}

type Sender struct {
	ID   uuid.UUID
	Role enums.SenderType // rider, driver, admin
}

type TraceInfo struct {
	RequestID  uuid.UUID
	Timestamp  int64
	RetryCount int
}

type LocationInfo struct {
	CountryCode string
}

type ClientInfo struct {
	AppVersion string // Added: For compatibility checks
	OS         string // Added: iOS/Android/Web
	Network    string // Added: 5G/WiFi
	DeviceID   string
}

type contextKey struct{}
type ContextManager struct{}

func NewContextManager() *ContextManager {
	return &ContextManager{}
}

func (m *ContextManager) Extract(ctx context.Context) (*RequestInfo, bool) {
	val, ok := ctx.Value(contextKey{}).(*RequestInfo)
	if !ok {
		return nil, false
	}
	return val, true
}

func (m *ContextManager) Inject(ctx context.Context, info *RequestInfo) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}
