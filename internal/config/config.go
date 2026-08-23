package config

import "time"

type Config struct {
	StationID             string
	SlotCount          int
	DefaultFlowLPM     float64
	FlowTolerancePct   float64
	PressureHoldMinutes int
	BoosterMinRun   time.Duration
	BoosterCoast    time.Duration
	ManifoldPrimeSec   int
	AlarmBufferSize    int
	ProcessTickMs      int
}

func Default() Config {
	return Config{
		StationID: "station-ss1", SlotCount: 8, DefaultFlowLPM: 12.5, FlowTolerancePct: 5,
		PressureHoldMinutes: 1, BoosterMinRun: time.Millisecond, BoosterCoast: time.Second,
		ManifoldPrimeSec: 5, AlarmBufferSize: 64, ProcessTickMs: 10,
	}
}

func (c Config) Validate() error {
	if c.SlotCount <= 0 {
		return errConfig("slot_count must be positive")
	}
	if c.DefaultFlowLPM < 0 {
		return errConfig("default_flow_lpm invalid")
	}
	return nil
}

func (c Config) ProcessTick() time.Duration {
	if c.ProcessTickMs <= 0 {
		return 100 * time.Millisecond
	}
	return time.Duration(c.ProcessTickMs) * time.Millisecond
}