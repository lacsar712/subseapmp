package app

import (
	"context"

	"github.com/lacsar712/subseapmp/internal/model"
)

func (a *App) BeginPumpScope(ctx context.Context, column model.PumpColumnID) (context.Context, context.CancelFunc) {
	if column == "" {
		column = model.PumpColumnID("col-primary")
	}
	a.pumpMu.Lock()
	if cancel, ok := a.pumpCancels[column]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.pumpCancels[column] = cancel
	a.pumpMu.Unlock()
	release := func() {
		a.pumpMu.Lock()
		delete(a.pumpCancels, column)
		a.pumpMu.Unlock()
		cancel()
	}
	return child, release
}

type PumpCycleOptions struct {
	Column model.PumpColumnID
}

func (a *App) RunPumpCycle(ctx context.Context, opt PumpCycleOptions) error {
	column := opt.Column
	if column == "" {
		column = model.PumpColumnID("col-primary")
	}
	cycleCtx, release := a.BeginPumpScope(ctx, column)
	defer release()
	if err := a.pumpFSM.Apply(cycleCtx, "arm"); err != nil {
		return err
	}
	return a.PumpStart(cycleCtx)
}
