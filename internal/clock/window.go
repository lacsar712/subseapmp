package clock

import (
	"fmt"
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

type InertWindow struct {
	clk      Clock
	duration time.Duration
}

func NewInertWindow(clk Clock, duration time.Duration) *InertWindow {
	if duration <= 0 {
		duration = 5 * time.Minute
	}
	return &InertWindow{clk: clk, duration: duration}
}

func (w *InertWindow) Ready(startedAt time.Time) bool {
	now := w.clk.Now()
	return !now.Before(startedAt.Add(w.duration))
}

func (w *InertWindow) Require(startedAt time.Time) error {
	if w.Ready(startedAt) {
		return nil
	}
	return fmt.Errorf("inert window: %w", model.ErrInertHold)
}
