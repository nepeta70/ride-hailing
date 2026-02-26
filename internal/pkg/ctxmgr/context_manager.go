package ctxmgr

import (
	"context"
)

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
