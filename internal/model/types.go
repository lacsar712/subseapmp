package model

import "time"

type StationState string

const (
	StationIdle      StationState = "idle"
	StationPriming   StationState = "priming"
	StationBoosting StationState = "circulate"
	StationHold      StationState = "hold"
	StationFault     StationState = "fault"
	StationShutdown  StationState = "shutdown"
)

type BoosterState string

const (
	BoosterOff     BoosterState = "off"
	BoosterStaging BoosterState = "staging"
	BoosterRun     BoosterState = "run"
	BoosterCoast   BoosterState = "coast"
	BoosterTrip    BoosterState = "trip"
)

type ValvePosition string

const (
	ValveClosed    ValvePosition = "closed"
	ValveOpen      ValvePosition = "open"
	ValveThrottled ValvePosition = "throttled"
)

type FlowSetpoint struct {
	LitersPerMinute float64
	TolerancePct    float64
}

func (f FlowSetpoint) Within(actual float64) bool {
	if f.LitersPerMinute <= 0 {
		return actual <= 0
	}
	lo := f.LitersPerMinute * (1 - f.TolerancePct/100)
	hi := f.LitersPerMinute * (1 + f.TolerancePct/100)
	return actual >= lo && actual <= hi
}

type PressureReading struct {
	Sensor  SensorID
	BarGauge float64
	At      time.Time
}

type SlotAssignment struct {
	Slot     SlotID
	Manifold ManifoldID
	Setpoint FlowSetpoint
	Enabled  bool
	LastFlow float64
}

type StationSnapshot struct {
	ID          StationID
	State       StationState
	Slots       []SlotAssignment
	Boosters []BoosterID
	UpdatedAt   time.Time
}

type BoostScheduleEntry struct {
	ID          ScheduleID
	Manifold    ManifoldID
	Start       time.Time
	End         time.Time
	Setpoint    FlowSetpoint
	HoldBar float64
}

type BoostSchedule struct {
	ID      ScheduleID
	Entries []BoostScheduleEntry
	Version int64
}

func (s BoostSchedule) Clone() BoostSchedule {
	out := BoostSchedule{ID: s.ID, Version: s.Version}
	if len(s.Entries) == 0 {
		return out
	}
	out.Entries = make([]BoostScheduleEntry, len(s.Entries))
	copy(out.Entries, s.Entries)
	return out
}

type AlarmEvent struct {
	Code     AlarmCode
	Message  string
	Station  StationID
	RaisedAt time.Time
	Severity int
}

type ManifoldRoute struct {
	From     ManifoldID
	To       ManifoldID
	Valve    ValveID
	Priority int
}