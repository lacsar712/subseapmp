package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/subseapmp/internal/model"
)

var CalibrateProbe func(ctx context.Context) error

func (a *App) CalibrateManifold(ctx context.Context, segment model.ManifoldID, holder string) error {
	lease, err := a.manifoldLeases.Acquire(string(segment), holder, a.clk.Now())
	if err != nil {
		return err
	}
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	lease.Release()
	return nil
}

func (a *App) ManifoldHeld(segment model.ManifoldID) bool {
	return a.manifoldLeases.IsHeld(string(segment))
}
