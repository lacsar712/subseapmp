package pump

import (
	"context"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/manifold"
	"github.com/lacsar712/subseapmp/internal/model"
	"github.com/lacsar712/subseapmp/internal/store"
	"github.com/lacsar712/subseapmp/internal/pipeline"
)

type BoosterPlant struct {
	cfg         config.Config
	clk         clock.Clock
	coordinator *BoosterCoordinator
	manifolds   *manifold.Registry
	flow        map[model.ManifoldID]*FlowController
	hold        *pipeline.HoldBarontroller
	sensors     *pipeline.SensorBank
	store       *store.Memory
}

func NewBoosterPlant(cfg config.Config, clk clock.Clock, mem *store.Memory) *BoosterPlant {
	return &BoosterPlant{
		cfg: cfg, clk: clk, coordinator: NewBoosterCoordinator(cfg, clk),
		manifolds: manifold.NewRegistry(), flow: make(map[model.ManifoldID]*FlowController),
		hold: pipeline.NewHoldBarontroller(clk), sensors: pipeline.NewSensorBank(clk), store: mem,
	}
}

func (p *BoosterPlant) PrimeManifold(ctx context.Context, id model.ManifoldID) error {
	m, ok := p.manifolds.Get(id)
	if !ok {
		return model.Wrap("pump", "manifold", model.ErrNotFound)
	}
	deadline := p.clk.Now().Add(time.Duration(p.cfg.ManifoldPrimeSec) * time.Second)
	primed := false
	for p.clk.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return model.Wrap("pump", "prime", context.Cause(ctx))
		default:
		}
		if !primed {
			m.Prime()
			primed = true
		}
		if m.Ready() {
			return nil
		}
		if pc, ok := p.clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return model.Wrap("pump", "prime_timeout", model.ErrConflict)
}

func (p *BoosterPlant) BindFlow(id model.ManifoldID, sp model.FlowSetpoint) { p.flow[id] = NewFlowController(sp) }

func (p *BoosterPlant) ObserveFlow(id model.ManifoldID, lpm float64) error {
	fc, ok := p.flow[id]
	if !ok {
		return model.Wrap("pump", "flow_bind", model.ErrNotFound)
	}
	fc.Observe(lpm)
	return nil
}

func (p *BoosterPlant) ValidateFlows(ctx context.Context) error {
	for id, fc := range p.flow {
		if err := fc.Validate(ctx); err != nil {
			return model.Wrap("pump", string(id), err)
		}
	}
	return nil
}

func (p *BoosterPlant) ArmPressureHold(w pipeline.Window) { p.hold.Arm(w) }
func (p *BoosterPlant) PressureHoldActive() bool                { return p.hold.Active() }
func (p *BoosterPlant) Coordinator() *BoosterCoordinator { return p.coordinator }
func (p *BoosterPlant) Manifolds() *manifold.Registry   { return p.manifolds }
func (p *BoosterPlant) Sensors() *pipeline.SensorBank      { return p.sensors }