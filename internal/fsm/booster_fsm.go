package fsm

import (
	"context"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

type BoosterSideEffect func(ctx context.Context, id model.BoosterID, from, to model.BoosterState) error

type BoosterFSM struct {
	clk      clock.Clock
	state    model.BoosterState
	id       model.BoosterID
	onChange BoosterSideEffect
	runSince time.Time
}

func NewBoosterFSM(clk clock.Clock, id model.BoosterID, effect BoosterSideEffect) *BoosterFSM {
	return &BoosterFSM{clk: clk, state: model.BoosterOff, id: id, onChange: effect}
}

func (f *BoosterFSM) State() model.BoosterState { return f.state }

func (f *BoosterFSM) Apply(ctx context.Context, event string) error {
	next, ok := AllowedBooster(f.state, event)
	if !ok {
		return model.Wrap("booster_fsm", "denied", model.ErrConflict)
	}
	prev := f.state
	if err := f.fire(ctx, prev, next); err != nil {
		return err
	}
	f.state = next
	if next == model.BoosterRun {
		f.runSince = f.clk.Now()
	}
	return nil
}

func (f *BoosterFSM) fire(ctx context.Context, from, to model.BoosterState) error {
	if f.onChange == nil {
		return nil
	}
	return f.onChange(ctx, f.id, from, to)
}

func (f *BoosterFSM) RunDuration() time.Duration {
	if f.state != model.BoosterRun {
		return 0
	}
	return f.clk.Now().Sub(f.runSince)
}

func (f *BoosterFSM) CanStop(minRun time.Duration) bool {
	if f.state != model.BoosterRun {
		return true
	}
	return f.RunDuration() >= minRun
}