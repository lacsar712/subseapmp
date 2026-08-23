package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/subseapmp/internal/model"
)

func TestCase(t *testing.T) {
	PumpBoostPulse = nil
	var pulses int
	PumpBoostPulse = func() { pulses++ }
	defer func() { PumpBoostPulse = nil }()
	f := NewPumpFSM(model.PumpColumnID("col-test"), nil)
	RegisterPumpBoostHook(f.Hooks())
	if err := f.Apply(context.Background(), "start"); err == nil {
		t.Fatal("expected illegal transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal transition should not pulse pump boost, got %d", pulses)
	}
}
