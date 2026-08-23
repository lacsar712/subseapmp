package app

import (
	"context"

	"github.com/lacsar712/subseapmp/internal/manifold"
	"github.com/lacsar712/subseapmp/internal/model"
	"github.com/lacsar712/subseapmp/internal/pipeline"
	"github.com/lacsar712/subseapmp/internal/pump"
)

func (a *App) initDomainModules(manifoldID model.ManifoldID) {
	headers := manifold.NewHeaderTable()
	hdr, _ := manifold.NewHeader("hdr-a", manifoldID, "v-intake", 10)
	headers.Add(hdr)
	a.slotCoord = manifold.NewSlotCoordinator(a.slots, headers)
	a.headers = headers

	a.segments = pipeline.NewSegmentRegistry()
	seg, _ := pipeline.NewSegment("seg-primary", manifoldID, "mf-secondary", 1200, 150, 300)
	seg.ObservePressure(6.5)
	_ = a.segments.Register(seg)
	a.routeResolver = pipeline.NewRouteResolver(a.segments, a.router)

	a.pressureBank = pump.NewBoosterPressureBank()
	sp := model.DefaultPressureSetpoint()
	ctrl := pump.NewPressureController(a.clk, sp)
	a.pressureBank.Bind(model.BoosterID("boost-1"), ctrl)
	a.staging = pump.NewStagingSequence(a.clk, a.plant.Coordinator(), a.pressureBank)
	a.staging.SetOrder(model.BoosterID("boost-1"))

	suct, _ := model.ParseSensorID("suct-1")
	disc, _ := model.ParseSensorID("disc-1")
	a.suctionGuard = pump.NewSuctionDischargeGuard(a.clk, a.plant.Sensors(), a.pressureBank, suct, disc, sp)
	a.suctionGuard.SeedSensors(45, 120)
	_ = a.suctionGuard.Observe(model.BoosterID("boost-1"))

	a.holdMonitor = pipeline.NewHoldMonitor(a.clk, a.plant.HoldController(), a.segments, a.plant.Sensors())
}

// CoordinateSlots enables slot zero and syncs header allocations.
func (a *App) CoordinateSlots(ctx context.Context) error {
	sp := model.FlowSetpoint{LitersPerMinute: a.cfg.DefaultFlowLPM, TolerancePct: a.cfg.FlowTolerancePct}
	if err := a.slotCoord.Assign(0, model.ManifoldID("mf-primary"), sp); err != nil {
		return err
	}
	if err := a.slotCoord.Enable(0, true); err != nil {
		return err
	}
	if err := a.slotCoord.SyncHeaders(); err != nil {
		return err
	}
	return a.slotCoord.ValidateSlotAlignment()
}

// ResolveManifoldRoute finds a viable segment route between manifolds.
func (a *App) ResolveManifoldRoute(from, to model.ManifoldID) (pipeline.ResolvedRoute, error) {
	route, err := a.routeResolver.Resolve(from, to)
	if err != nil {
		return pipeline.ResolvedRoute{}, err
	}
	if err := a.routeResolver.SegmentPressureCheck(route); err != nil {
		return pipeline.ResolvedRoute{}, err
	}
	return route, nil
}

// StageBoosters runs the staged booster ramp-up sequence.
func (a *App) StageBoosters(ctx context.Context) error {
	if err := a.suctionGuard.Observe(model.BoosterID("boost-1")); err != nil {
		return err
	}
	if err := a.suctionGuard.ValidateAll(ctx); err != nil {
		return err
	}
	return a.staging.StageBatch(ctx)
}

// ValidateSuctionDischarge checks intake/discharge for a booster unit.
func (a *App) ValidateSuctionDischarge(ctx context.Context, id model.BoosterID) error {
	return a.suctionGuard.Validate(ctx, id)
}

// MonitorPressureHold ticks the hold monitor during an active window.
func (a *App) MonitorPressureHold(ctx context.Context) error {
	return a.holdMonitor.Tick(ctx)
}
