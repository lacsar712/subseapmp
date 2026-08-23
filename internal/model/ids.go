package model

import (
	"fmt"
	"strings"
)

type StationID string
type SlotID string
type ManifoldID string
type BoosterID string
type ValveID string
type SensorID string
type ScheduleID string
type AlarmCode string

func (id StationID) String() string       { return string(id) }
func (id SlotID) String() string       { return string(id) }
func (id ManifoldID) String() string   { return string(id) }
func (id BoosterID) String() string { return string(id) }
func (id ValveID) String() string      { return string(id) }
func (id SensorID) String() string     { return string(id) }
func (id ScheduleID) String() string   { return string(id) }
func (id AlarmCode) String() string    { return string(id) }

func ParseStationID(raw string) (StationID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return StationID(raw), nil
}

func ParseSlotID(rack StationID, index int) (SlotID, error) {
	if rack == "" || index < 0 {
		return "", ErrInvalidID
	}
	return SlotID(fmt.Sprintf("%s-slot-%02d", rack, index)), nil
}

func ParseManifoldID(raw string) (ManifoldID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ManifoldID(raw), nil
}

func ParseBoosterID(raw string) (BoosterID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return BoosterID(raw), nil
}

func ParseValveID(raw string) (ValveID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ValveID(raw), nil
}

func ParseSensorID(raw string) (SensorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return SensorID(raw), nil
}

func ParseScheduleID(raw string) (ScheduleID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ScheduleID(raw), nil
}