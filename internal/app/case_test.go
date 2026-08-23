package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	anchor := a.Clock().Now()
	entries := []model.BoostScheduleEntry{{
		Manifold: model.ManifoldID("mf-primary"),
		Start:    anchor, End: anchor.Add(time.Hour),
		Setpoint: model.FlowSetpoint{LitersPerMinute: 12.5, TolerancePct: 5},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = a.ExecutePlan(ctx, entries)
	if a.SchedulerItemCount() > 0 {
		t.Fatal("cancelled ctx still installed boost plan items")
	}
}
