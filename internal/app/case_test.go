package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/subseapmp/internal/config"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.PumpStart(ctx)
	}()
	time.Sleep(15 * time.Millisecond)
	cancel()
	<-done
	if a.PumpRampPressure() > 10 {
		t.Fatalf("pump ramp continued after cancel, pressure=%.1f", a.PumpRampPressure())
	}
}
