package fsm

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestStationFSMPrime(t *testing.T) {
	f := NewStationFSM("r1", nil)
	if err := f.Apply(context.Background(), "prime"); err != nil {
		t.Fatal(err)
	}
	if f.State() != model.StationPriming {
		t.Fatalf("state %s", f.State())
	}
}

func TestBoosterFSMRun(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	f := NewBoosterFSM(clk, "c1", func(ctx context.Context, id model.BoosterID, from, to model.BoosterState) error {
		return nil
	})
	_ = f.Apply(context.Background(), "start")
	_ = f.Apply(context.Background(), "staged")
	if f.State() != model.BoosterRun {
		t.Fatal("booster run")
	}
}