package fsm

import (
	"context"
	"fmt"

	"github.com/lacsar712/subseapmp/internal/model"
)

type PumpState string

const (
	PumpIdle     PumpState = "idle"
	PumpArming   PumpState = "arming"
	PumpRunning  PumpState = "running"
	PumpCoast    PumpState = "coast"
)

type PumpFSM struct {
	id    model.PumpColumnID
	state PumpState
	hooks *PumpHookChain
}

func NewPumpFSM(id model.PumpColumnID, effect func(context.Context, model.PumpColumnID, PumpState, PumpState) error) *PumpFSM {
	_ = effect
	return &PumpFSM{id: id, state: PumpIdle, hooks: NewPumpHookChain()}
}

func (f *PumpFSM) Hooks() *PumpHookChain { return f.hooks }

func (f *PumpFSM) State() PumpState { return f.state }

func (f *PumpFSM) Apply(ctx context.Context, event string) error {
	next, ok := allowedPump(f.state, event)
	if !ok {
		// Illegal transition: state is unchanged, so no after-hooks (which would
		// fire side effects such as the boost pulse) may run. Matches BoosterFSM
		// and StationFSM, which short-circuit on a denied transition.
		return fmt.Errorf("%s from %s: %w", event, f.state, model.ErrConflict)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return err
		}
	}
	return nil
}

func allowedPump(from PumpState, event string) (PumpState, bool) {
	switch from {
	case PumpIdle:
		if event == "arm" {
			return PumpArming, true
		}
	case PumpArming:
		if event == "start" {
			return PumpRunning, true
		}
	case PumpRunning:
		if event == "coast" {
			return PumpCoast, true
		}
	case PumpCoast:
		if event == "done" {
			return PumpIdle, true
		}
	}
	return from, false
}
