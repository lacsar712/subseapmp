package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestRunOnce(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.StationFSM.State() != model.StationBoosting {
		t.Fatalf("state %s", a.StationFSM.State())
	}
}

func TestApplyScheduleSnapshot(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	now := a.clk.Now()
	a.sched.Save(model.BoostSchedule{ID: "sch1", Entries: []model.BoostScheduleEntry{{
		Start: now.Add(-time.Hour), End: now.Add(time.Hour), Manifold: "mf-primary",
		Setpoint: model.FlowSetpoint{LitersPerMinute: 8, TolerancePct: 5}, HoldBar: 5,
	}}})
	if err := a.ApplyScheduleSnapshot(context.Background(), "sch1"); err != nil {
		t.Fatal(err)
	}
}