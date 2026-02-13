package ctxmgr

import "context"

type RequestInfo struct {
	User     UserSession
	Trace    TraceInfo
	Location LocationInfo
	Client   ClientInfo
}

type UserSession struct {
	ID   string
	Role string // rider, driver, admin
}

type TraceInfo struct {
	RequestID  string
	Origin     string
	Timestamp  string
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
	val, ok := ctx.Value(contextKey{}).(RequestInfo)
	if !ok {
		return nil, false
	}
	return &val, true
}

func (m *ContextManager) Inject(ctx context.Context, info RequestInfo) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}
