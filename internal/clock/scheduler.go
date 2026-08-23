package clock

import (
	"context"
	"sync"

	"github.com/lacsar712/subseapmp/internal/model"
)

type PumpScheduler struct {
	mu    sync.Mutex
	items []string
	clk   Clock
}

func NewPumpScheduler(clk Clock) *PumpScheduler {
	return &PumpScheduler{clk: clk}
}

func (s *PumpScheduler) ItemCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

func (s *PumpScheduler) InstallBoostPlanCtx(ctx context.Context, entries []model.BoostScheduleEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		s.items = append(s.items, string(e.Manifold)+":"+e.Start.String())
		s.mu.Unlock()
	}
	return nil
}
