package store

import (
	"sync"

	"github.com/lacsar712/subseapmp/internal/model"
)

type Memory struct {
	mu        sync.RWMutex
	racks     map[model.StationID]model.StationSnapshot
	schedules map[model.ScheduleID]model.BoostSchedule
}

func NewMemory() *Memory {
	return &Memory{
		racks:     make(map[model.StationID]model.StationSnapshot),
		schedules: make(map[model.ScheduleID]model.BoostSchedule),
	}
}

func (m *Memory) PutStation(snap model.StationSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.racks[snap.ID] = snap
}

func (m *Memory) GetStation(id model.StationID) (model.StationSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.racks[id]
	return s, ok
}

func (m *Memory) PutSchedule(s model.BoostSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schedules[s.ID] = s
}

func (m *Memory) GetSchedule(id model.ScheduleID) (model.BoostSchedule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	return s, ok
}

func (m *Memory) ListStations() []model.StationSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.StationSnapshot, 0, len(m.racks))
	for _, v := range m.racks {
		out = append(out, v)
	}
	return out
}