package silo

import (
	"context"
	"sync"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type GrainActivation struct {
	identity *grain.GrainIdentity
	instance ports.Grain
	mu       sync.Mutex
}

func (a *GrainActivation) Identity() *grain.GrainIdentity {
	return a.identity
}

func (a *GrainActivation) OnReceive(ctx context.Context, msg ports.Message) (ports.Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.instance.OnReceive(ctx, msg)
}

func (a *GrainActivation) deactivate(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.instance.OnDeactivate(ctx)
}

func (a *GrainActivation) GetStatus() any {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.instance.GetStatus()
}

func (a *GrainActivation) IsTerminal() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.instance.IsTerminal()
}

var _ ports.GrainRef = (*GrainActivation)(nil)
