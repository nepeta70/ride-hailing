package grains

import (
	"context"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RidePipeline struct {
	g    *RideGrain
	ctx  context.Context
	span trace.Span
	err  error
}

func (g *RideGrain) Start(ctx context.Context, name string) *RidePipeline {
	ctx, span := g.telemetry.Tracer().Start(ctx, name)
	return &RidePipeline{g: g, ctx: ctx, span: span}
}

func (p *RidePipeline) Step(name string, fn func(ctx context.Context) error) *RidePipeline {
	if p.err != nil {
		return p
	}

	ctx, stepSpan := p.g.telemetry.Tracer().Start(p.ctx, name)
	defer stepSpan.End()

	p.err = fn(ctx)

	p.ctx = ctx

	if p.err != nil {
		stepSpan.RecordError(p.err)
		stepSpan.SetStatus(codes.Error, p.err.Error())
	}
	return p
}

func (p *RidePipeline) End(resp pkgPorts.Message) (pkgPorts.Message, error) {
	defer p.span.End()
	if p.err != nil {
		p.span.SetAttributes(attribute.Bool("message.handled", false))
		p.span.SetStatus(codes.Error, p.err.Error())
		return nil, p.err
	} else {
		p.span.SetStatus(codes.Ok, "success")
		p.span.SetAttributes(attribute.Bool("message.handled", true))
		if p.err == nil {
			p.span.SetAttributes(attribute.String("grain.newstatus", p.g.state.Status.String()))
			p.span.SetAttributes(attribute.Int("grain.newversion", p.g.version))
		}
	}
	return resp, nil
}
