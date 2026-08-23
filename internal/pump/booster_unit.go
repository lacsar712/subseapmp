package pump

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/fsm"
	"github.com/lacsar712/subseapmp/internal/model"
)

type BoosterCoordinator struct {
	mu    sync.Mutex
	cfg   config.Config
	clk   clock.Clock
	units map[model.BoosterID]*fsm.BoosterFSM
	log   []string
}

func NewBoosterCoordinator(cfg config.Config, clk clock.Clock) *BoosterCoordinator {
	return &BoosterCoordinator{cfg: cfg, clk: clk, units: make(map[model.BoosterID]*fsm.BoosterFSM)}
}

func (c *BoosterCoordinator) Register(id model.BoosterID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.units[id]; ok {
		return
	}
	effect := func(ctx context.Context, cid model.BoosterID, from, to model.BoosterState) error {
		c.log = append(c.log, fmt.Sprintf("%s %s->%s", cid, from, to))
		if to == model.BoosterTrip {
			return model.Wrap("booster", "trip", model.ErrBooster)
		}
		return nil
	}
	c.units[id] = fsm.NewBoosterFSM(c.clk, id, effect)
}

func (c *BoosterCoordinator) Start(ctx context.Context, id model.BoosterID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("booster", "missing", model.ErrNotFound)
	}
	if err := unit.Apply(ctx, "start"); err != nil {
		return err
	}
	return unit.Apply(ctx, "staged")
}

func (c *BoosterCoordinator) Stop(ctx context.Context, id model.BoosterID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("booster", "missing", model.ErrNotFound)
	}
	if !unit.CanStop(c.cfg.BoosterMinRun) {
		return model.Wrap("booster", "min_run", model.ErrConflict)
	}
	if err := unit.Apply(ctx, "stop"); err != nil {
		return err
	}
	return unit.Apply(ctx, "coast_done")
}

func (c *BoosterCoordinator) States() map[model.BoosterID]model.BoosterState {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[model.BoosterID]model.BoosterState, len(c.units))
	for id, u := range c.units {
		out[id] = u.State()
	}
	return out
}