package app

import (
	"context"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
)

type PumpRamp struct {
	clk   clock.Clock
	steps int
	delay time.Duration
}

func NewPumpRamp(clk clock.Clock, steps int, delay time.Duration) *PumpRamp {
	if steps <= 0 {
		steps = 40
	}
	if delay <= 0 {
		delay = time.Millisecond
	}
	return &PumpRamp{clk: clk, steps: steps, delay: delay}
}

func (r *PumpRamp) Ramp(ctx context.Context, apply func(float64)) error {
	target := 100.0
	step := target / float64(r.steps)
	cur := 0.0
	for cur < target {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		cur += step
		if cur > target {
			cur = target
		}
		apply(cur)
		if pc, ok := r.clk.(*clock.ProcessClock); ok {
			pc.Step()
		}
		time.Sleep(r.delay)
	}
	return nil
}

func (a *App) PumpStart(ctx context.Context) error {
	return a.pumpRamp.Ramp(ctx, func(v float64) { a.rampPressure = v })
}

func (a *App) PumpRampPressure() float64 { return a.rampPressure }

func (a *App) Clock() clock.Clock { return a.clk }
