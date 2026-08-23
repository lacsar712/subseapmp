package pump

import (
	"context"
	"fmt"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
	"github.com/lacsar712/subseapmp/internal/pipeline"
)

// SuctionDischargeGuard validates intake/discharge pressures before flow changes.
type SuctionDischargeGuard struct {
	clk      clock.Clock
	sensors  *pipeline.SensorBank
	bank     *BoosterPressureBank
	suction  model.SensorID
	discharge model.SensorID
	setpoint model.PressureSetpoint
}

func NewSuctionDischargeGuard(
	clk clock.Clock,
	sensors *pipeline.SensorBank,
	bank *BoosterPressureBank,
	suction, discharge model.SensorID,
	sp model.PressureSetpoint,
) *SuctionDischargeGuard {
	return &SuctionDischargeGuard{
		clk: clk, sensors: sensors, bank: bank,
		suction: suction, discharge: discharge, setpoint: sp,
	}
}

// Observe reads sensor values and updates the pressure bank for a booster.
func (g *SuctionDischargeGuard) Observe(booster model.BoosterID) error {
	suct, ok := g.sensors.Reading(g.suction)
	if !ok {
		return model.Wrap("suction_guard", string(g.suction), model.ErrNotFound)
	}
	disc, ok := g.sensors.Reading(g.discharge)
	if !ok {
		return model.Wrap("suction_guard", string(g.discharge), model.ErrNotFound)
	}
	ctrl, ok := g.bank.Get(booster)
	if !ok {
		ctrl = NewPressureController(g.clk, g.setpoint)
		g.bank.Bind(booster, ctrl)
	}
	ctrl.Observe(suct.BarGauge, disc.BarGauge)
	return nil
}

// Validate checks suction floor, discharge ceiling, and differential band.
func (g *SuctionDischargeGuard) Validate(ctx context.Context, booster model.BoosterID) error {
	if err := g.Observe(booster); err != nil {
		return err
	}
	ctrl, ok := g.bank.Get(booster)
	if !ok {
		return model.Wrap("suction_guard", string(booster), model.ErrNotFound)
	}
	if err := ctrl.Validate(ctx); err != nil {
		return err
	}
	suct, _ := g.sensors.Reading(g.suction)
	disc, _ := g.sensors.Reading(g.discharge)
	if suct.BarGauge < g.setpoint.MinSuctionBar {
		return model.Wrap("suction_guard", "low_suction",
			fmt.Errorf("%.1f bar below %.1f minimum", suct.BarGauge, g.setpoint.MinSuctionBar))
	}
	if disc.BarGauge > g.setpoint.MaxDischargeBar {
		return model.Wrap("suction_guard", "high_discharge",
			fmt.Errorf("%.1f bar above %.1f maximum", disc.BarGauge, g.setpoint.MaxDischargeBar))
	}
	return nil
}

// ValidateAll runs suction/discharge checks for every bound booster.
func (g *SuctionDischargeGuard) ValidateAll(ctx context.Context) error {
	return g.bank.ValidateAll(ctx)
}

// Differential returns the current suction-to-discharge delta for a booster.
func (g *SuctionDischargeGuard) Differential(booster model.BoosterID) (float64, error) {
	ctrl, ok := g.bank.Get(booster)
	if !ok {
		return 0, model.Wrap("suction_guard", string(booster), model.ErrNotFound)
	}
	return ctrl.Differential(), nil
}

// WithinBand reports whether a booster's differential is inside the setpoint band.
func (g *SuctionDischargeGuard) WithinBand(booster model.BoosterID) bool {
	ctrl, ok := g.bank.Get(booster)
	if !ok {
		return false
	}
	return ctrl.WithinSetpoint()
}

// SeedSensors primes sensor readings for validation during manifold prime.
func (g *SuctionDischargeGuard) SeedSensors(suctionBar, dischargeBar float64) {
	g.sensors.Set(g.suction, suctionBar)
	g.sensors.Set(g.discharge, dischargeBar)
}
