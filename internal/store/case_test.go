package store

import (
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

func TestCase(t *testing.T) {
	orig := PressureSnapshot{
		StationID: model.StationID("station-ss1"),
		BoostSlots: []model.BoostScheduleEntry{{
			Manifold: model.ManifoldID("mf-primary"),
			Start:    time.Unix(0, 0), End: time.Unix(100, 0),
			Setpoint: model.FlowSetpoint{LitersPerMinute: 12.5, TolerancePct: 5},
		}},
	}
	clone := orig.Clone()
	clone.BoostSlots[0].Setpoint.LitersPerMinute = 99
	if orig.BoostSlots[0].Setpoint.LitersPerMinute == 99 {
		t.Fatal("clone mutated original BoostSlots backing array")
	}
}
