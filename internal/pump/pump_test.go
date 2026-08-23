package pump

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/interlock"
	"github.com/lacsar712/subseapmp/internal/manifold"
	"github.com/lacsar712/subseapmp/internal/model"
	"github.com/lacsar712/subseapmp/internal/store"
)

func TestPumpPrimeAndFlow(t *testing.T) {
	cfg := config.Default()
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	plant := NewBoosterPlant(cfg, clk, store.NewMemory())
	plant.Manifolds().Add(&manifold.Manifold{ID: "mf1"})
	plant.BindFlow("mf1", model.FlowSetpoint{LitersPerMinute: 10, TolerancePct: 10})
	if err := plant.PrimeManifold(context.Background(), "mf1"); err != nil {
		t.Fatal(err)
	}
	_ = plant.ObserveFlow("mf1", 10)
	if err := plant.ValidateFlows(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestValveActuator(t *testing.T) {
	v := NewValveActuator(interlock.NewValveLock(time.Now))
	if err := v.Open(context.Background(), "v1", 1); err != nil {
		t.Fatal(err)
	}
	if v.Position("v1") != model.ValveOpen {
		t.Fatal("open")
	}
}