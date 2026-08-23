package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/subseapmp/internal/alarms"
	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/pump"
	"github.com/lacsar712/subseapmp/internal/fsm"
	"github.com/lacsar712/subseapmp/internal/interlock"
	"github.com/lacsar712/subseapmp/internal/manifold"
	"github.com/lacsar712/subseapmp/internal/model"
	"github.com/lacsar712/subseapmp/internal/store"
	"github.com/lacsar712/subseapmp/internal/pipeline"
)

type App struct {
	cfg     config.Config
	clk     clock.Clock
	mem     *store.Memory
	sched   *store.ScheduleStore
	plant   *pump.BoosterPlant
	StationFSM *fsm.StationFSM
	slots   *manifold.SlotTable
	alarms  *alarms.Emitter
	lock    *interlock.ValveLock
	valves  *pump.ValveActuator
	router  *manifold.Router
	slotCoord    *manifold.SlotCoordinator
	headers      *manifold.HeaderTable
	segments     *pipeline.SegmentRegistry
	routeResolver *pipeline.RouteResolver
	pressureBank *pump.BoosterPressureBank
	staging      *pump.StagingSequence
	suctionGuard *pump.SuctionDischargeGuard
	holdMonitor     *pipeline.HoldMonitor
	pumpRamp        *PumpRamp
	pumpSched       *clock.PumpScheduler
	inertWindow     *clock.InertWindow
	manifoldLeases  *interlock.ManifoldLeaseRegistry
	leakGuard       *interlock.ManifoldLeakGuard
	pumpMu          sync.Mutex
	pumpCancels     map[model.PumpColumnID]context.CancelFunc
	pumpFSM         *fsm.PumpFSM
	rampPressure    float64
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, cfg.ProcessTick())
	mem := store.NewMemory()
	StationID, err := model.ParseStationID(cfg.StationID)
	if err != nil {
		return nil, err
	}
	manifoldID := model.ManifoldID("mf-primary")
	slots, err := manifold.NewSlotTable(StationID, cfg.SlotCount, manifoldID)
	if err != nil {
		return nil, err
	}
	plant := pump.NewBoosterPlant(cfg, clk, mem)
	plant.Manifolds().Add(&manifold.Manifold{ID: manifoldID, Capacity: 100})
	plant.BindFlow(manifoldID, model.FlowSetpoint{LitersPerMinute: cfg.DefaultFlowLPM, TolerancePct: cfg.FlowTolerancePct})
	plant.Coordinator().Register(model.BoosterID("boost-1"))
	a := &App{
		cfg: cfg, clk: clk, mem: mem, sched: store.NewScheduleStore(mem), plant: plant, slots: slots,
		lock: interlock.NewValveLock(clk.Now),
		router: manifold.NewRouter([]model.ManifoldRoute{{From: manifoldID, To: "mf-secondary", Valve: "v-main", Priority: 10}}),
	}
	a.valves = pump.NewValveActuator(a.lock)
	a.alarms = alarms.NewEmitter(alarms.NewRegistry(), clk, cfg.AlarmBufferSize)
	a.StationFSM = fsm.NewStationFSM(StationID, a.onStationTransition)
	a.initDomainModules(manifoldID)
	a.persistSnapshot(StationID)
	return a, nil
}

func (a *App) onStationTransition(ctx context.Context, rack model.StationID, from, to model.StationState) error {
	if to == model.StationFault {
		return a.alarms.Raise(ctx, "PUMP_TRIP", rack, 3)
	}
	return nil
}

func (a *App) onPumpTransition(ctx context.Context, column model.PumpColumnID, from, to fsm.PumpState) error {
	_ = ctx
	_ = column
	_ = from
	_ = to
	return nil
}

func (a *App) persistSnapshot(id model.StationID) {
	b := store.NewSnapshotBuilder(id).State(a.StationFSM.State())
	for _, s := range a.slots.Slots() {
		b.Slot(model.SlotAssignment{
			Slot: s.ID, Manifold: s.Manifold, Enabled: s.Enabled,
			Setpoint: model.FlowSetpoint{LitersPerMinute: a.cfg.DefaultFlowLPM, TolerancePct: a.cfg.FlowTolerancePct},
		})
	}
	a.mem.PutStation(b.Build(a.clk.Now()))
}

func (a *App) ApplyScheduleSnapshot(ctx context.Context, id model.ScheduleID) error {
	snap, err := a.sched.SnapshotClone(id)
	if err != nil {
		return err
	}
	now := a.clk.Now()
	entry, ok := a.sched.ActiveEntry(snap, now)
	if !ok {
		return model.Wrap("app", "schedule", model.ErrScheduleEmpty)
	}
	a.plant.BindFlow(entry.Manifold, entry.Setpoint)
	a.plant.ArmPressureHold(pipeline.NewWindow(now, time.Duration(a.cfg.PressureHoldMinutes)*time.Minute, entry.HoldBar))
	return nil
}

func (a *App) RunOnce(ctx context.Context) error {
	if err := a.StationFSM.Apply(ctx, "prime"); err != nil {
		return err
	}
	mf := model.ManifoldID("mf-primary")
	if err := a.plant.PrimeManifold(ctx, mf); err != nil {
		return err
	}
	if _, err := a.ResolveManifoldRoute(mf, "mf-secondary"); err != nil {
		return err
	}
	if err := a.CoordinateSlots(ctx); err != nil {
		return err
	}
	if err := a.StationFSM.Apply(ctx, "flow_ok"); err != nil {
		return err
	}
	if err := a.StageBoosters(ctx); err != nil {
		return err
	}
	a.plant.ObserveFlow(mf, a.cfg.DefaultFlowLPM)
	if err := a.plant.ValidateFlows(ctx); err != nil {
		return err
	}
	a.plant.ArmPressureHold(pipeline.NewWindow(a.clk.Now(), time.Duration(a.cfg.PressureHoldMinutes)*time.Minute, 6.5))
	a.holdMonitor.BeginHold(pipeline.NewWindow(a.clk.Now(), time.Duration(a.cfg.PressureHoldMinutes)*time.Minute, 6.5))
	if err := a.MonitorPressureHold(ctx); err != nil {
		return err
	}
	if err := a.StationFSM.Apply(ctx, "pressure_hold"); err != nil {
		return err
	}
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(time.Duration(a.cfg.PressureHoldMinutes)*time.Minute + time.Second)
	}
	if err := a.StationFSM.Apply(ctx, "release_hold"); err != nil {
		return err
	}
	a.persistSnapshot(model.StationID(a.cfg.StationID))
	return nil
}

func (a *App) StatusLine() string {
	return fmt.Sprintf("station=%s state=%s hold=%v slots=%d", a.cfg.StationID, a.StationFSM.State(), a.plant.PressureHoldActive(), a.slots.EnabledCount())
}