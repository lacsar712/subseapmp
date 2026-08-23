package pump

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

// PressureController tracks suction and discharge setpoints for multiphase boosters.
type PressureController struct {
	mu          sync.RWMutex
	clk         clock.Clock
	suctionBar  float64
	dischargeBar float64
	setpoint    model.PressureSetpoint
	alarms      []string
}

type PressureSetpoint = model.PressureSetpoint

func NewPressureController(clk clock.Clock, sp model.PressureSetpoint) *PressureController {
	return &PressureController{clk: clk, setpoint: sp}
}

func (p *PressureController) Observe(suction, discharge float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.suctionBar = suction
	p.dischargeBar = discharge
}

func (p *PressureController) Suction() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.suctionBar
}

func (p *PressureController) Discharge() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dischargeBar
}

func (p *PressureController) Differential() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dischargeBar - p.suctionBar
}

func (p *PressureController) WithinSetpoint() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	diff := p.dischargeBar - p.suctionBar
	return diff >= p.setpoint.MinDiffBar && diff <= p.setpoint.MaxDiffBar
}

func (p *PressureController) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return model.Wrap("pressure", "canceled", context.Cause(ctx))
	default:
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.suctionBar < p.setpoint.MinSuctionBar {
		return model.Wrap("pressure", "low_suction", fmt.Errorf("%.1f < %.1f", p.suctionBar, p.setpoint.MinSuctionBar))
	}
	if p.dischargeBar > p.setpoint.MaxDischargeBar {
		return model.Wrap("pressure", "high_discharge", fmt.Errorf("%.1f > %.1f", p.dischargeBar, p.setpoint.MaxDischargeBar))
	}
	diff := p.dischargeBar - p.suctionBar
	if diff < p.setpoint.MinDiffBar || diff > p.setpoint.MaxDiffBar {
		return model.Wrap("pressure", "diff_out_of_band", model.ErrConflict)
	}
	return nil
}

func (p *PressureController) RecordAlarm(code string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.alarms = append(p.alarms, code)
}

func (p *PressureController) Alarms() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]string, len(p.alarms))
	copy(out, p.alarms)
	return out
}

// BoosterPressureBank tracks one controller per booster unit on the station.
type BoosterPressureBank struct {
	mu    sync.RWMutex
	units map[model.BoosterID]*PressureController
}

func NewBoosterPressureBank() *BoosterPressureBank {
	return &BoosterPressureBank{units: make(map[model.BoosterID]*PressureController)}
}

func (b *BoosterPressureBank) Bind(id model.BoosterID, ctrl *PressureController) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.units[id] = ctrl
}

func (b *BoosterPressureBank) Get(id model.BoosterID) (*PressureController, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	c, ok := b.units[id]
	return c, ok
}

func (b *BoosterPressureBank) ValidateAll(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for id, ctrl := range b.units {
		if err := ctrl.Validate(ctx); err != nil {
			return model.Wrap("pressure_bank", string(id), err)
		}
	}
	return nil
}
