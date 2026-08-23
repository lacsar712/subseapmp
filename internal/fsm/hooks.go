package fsm

import (
	"context"
)

type PumpTransitionHook func(ctx context.Context, from, to PumpState, event string) error

type PumpHookChain struct {
	before []PumpTransitionHook
	after  []PumpTransitionHook
}

func NewPumpHookChain() *PumpHookChain { return &PumpHookChain{} }

func (h *PumpHookChain) OnBefore(fn PumpTransitionHook) { h.before = append(h.before, fn) }
func (h *PumpHookChain) OnAfter(fn PumpTransitionHook)  { h.after = append(h.after, fn) }

func (h *PumpHookChain) RunBefore(ctx context.Context, from, to PumpState, event string) error {
	for _, fn := range h.before {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

func (h *PumpHookChain) RunAfter(ctx context.Context, from, to PumpState, event string) error {
	for _, fn := range h.after {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

var PumpBoostPulse func()

func RegisterPumpBoostHook(chain *PumpHookChain) {
	chain.OnAfter(func(ctx context.Context, from, to PumpState, event string) error {
		_ = ctx
		_ = from
		_ = to
		_ = event
		if PumpBoostPulse != nil {
			PumpBoostPulse()
		}
		return nil
	})
}
