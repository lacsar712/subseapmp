package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/subseapmp/internal/model"
)

func (a *App) ReportLinePressureFault(ctx context.Context, bar float64) error {
	limit := 6.5
	if bar > limit {
		return fmt.Errorf("line pressure: %w", model.ErrLinePressureHigh)
	}
	return nil
}

func (a *App) HandleManifoldLeak(ctx context.Context, delta float64) error {
	if err := a.leakGuard.Permit(delta); err != nil {
		_ = a.alarms.Raise(ctx, "MANIFOLD_LEAK", model.StationID(a.cfg.StationID), 2)
		return fmt.Errorf("manifold: leak interlock active")
	}
	return nil
}
