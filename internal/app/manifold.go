package app

import (
	"context"
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

func (a *App) ExecutePlan(ctx context.Context, entries []model.BoostScheduleEntry) error {
	return a.pumpSched.InstallBoostPlanCtx(ctx, entries)
}

func (a *App) SchedulerItemCount() int {
	return a.pumpSched.ItemCount()
}

func (a *App) RunInertScheduler(ctx context.Context, anchor time.Time) error {
	if err := a.inertWindow.Require(anchor); err != nil {
		return err
	}
	entries := []model.BoostScheduleEntry{{
		Manifold: model.ManifoldID("mf-primary"),
		Start:    anchor, End: anchor.Add(time.Hour),
		Setpoint: model.FlowSetpoint{LitersPerMinute: a.cfg.DefaultFlowLPM, TolerancePct: a.cfg.FlowTolerancePct},
	}}
	return a.ExecutePlan(ctx, entries)
}
