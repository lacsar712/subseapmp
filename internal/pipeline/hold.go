package pipeline

import (
	"context"
	"errors"
	"sync"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

type HoldBarontroller struct {
	mu     sync.Mutex
	clk    clock.Clock
	window Window
	active bool
}

func NewHoldBarontroller(clk clock.Clock) *HoldBarontroller { return &HoldBarontroller{clk: clk} }

func (h *HoldBarontroller) Arm(w Window) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.window = w
	h.active = true
}

func (h *HoldBarontroller) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.active = false
}

func (h *HoldBarontroller) Active() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.active && h.window.Active(h.clk)
}

func (h *HoldBarontroller) WaitStable(ctx context.Context, readings <-chan model.PressureReading) error {
	for {
		select {
		case <-ctx.Done():
			return errors.Join(model.ErrContextCanceled, context.Cause(ctx))
		case r, ok := <-readings:
			if !ok {
				return model.ErrPressureHold
			}
			h.mu.Lock()
			w := h.window
			act := h.active
			h.mu.Unlock()
			if !act {
				return nil
			}
			if w.WithinHold(r) && !w.Active(h.clk) {
				return nil
			}
		}
	}
}