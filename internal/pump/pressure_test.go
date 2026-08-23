package pump

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestPressureControllerWithinSetpoint(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	ctrl := NewPressureController(clk, model.DefaultPressureSetpoint())
	ctrl.Observe(40, 120)
	if !ctrl.WithinSetpoint() {
		t.Fatal("expected within setpoint")
	}
}

func TestPressureControllerValidateLowSuction(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	ctrl := NewPressureController(clk, model.DefaultPressureSetpoint())
	ctrl.Observe(5, 100)
	if err := ctrl.Validate(context.Background()); err == nil {
		t.Fatal("expected low suction error")
	}
}

func TestBoosterPressureBankValidateAll(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	bank := NewBoosterPressureBank()
	c1 := NewPressureController(clk, model.DefaultPressureSetpoint())
	c1.Observe(50, 130)
	bank.Bind("boost-1", c1)
	if err := bank.ValidateAll(context.Background()); err != nil {
		t.Fatal(err)
	}
}
