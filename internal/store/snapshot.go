package store

import (
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

type SnapshotBuilder struct {
	id    model.StationID
	state model.StationState
	slots []model.SlotAssignment
	comp  []model.BoosterID
}

func NewSnapshotBuilder(id model.StationID) *SnapshotBuilder {
	return &SnapshotBuilder{id: id, state: model.StationIdle}
}

func (b *SnapshotBuilder) State(s model.StationState) *SnapshotBuilder {
	b.state = s
	return b
}

func (b *SnapshotBuilder) Slot(a model.SlotAssignment) *SnapshotBuilder {
	b.slots = append(b.slots, a)
	return b
}

func (b *SnapshotBuilder) Booster(id model.BoosterID) *SnapshotBuilder {
	b.comp = append(b.comp, id)
	return b
}

func (b *SnapshotBuilder) Build(at time.Time) model.StationSnapshot {
	slots := make([]model.SlotAssignment, len(b.slots))
	copy(slots, b.slots)
	comp := make([]model.BoosterID, len(b.comp))
	copy(comp, b.comp)
	return model.StationSnapshot{ID: b.id, State: b.state, Slots: slots, Boosters: comp, UpdatedAt: at}
}

type PressureSnapshot struct {
	StationID  model.StationID
	BoostSlots []model.BoostScheduleEntry
}

func (p PressureSnapshot) Clone() PressureSnapshot {
	out := PressureSnapshot{StationID: p.StationID}
	if len(p.BoostSlots) == 0 {
		return out
	}
	out.BoostSlots = make([]model.BoostScheduleEntry, len(p.BoostSlots))
	copy(out.BoostSlots, p.BoostSlots)
	return out
}

func DiffSlots(before, after model.StationSnapshot) []model.SlotID {
	index := make(map[model.SlotID]model.SlotAssignment)
	for _, s := range before.Slots {
		index[s.Slot] = s
	}
	var changed []model.SlotID
	for _, s := range after.Slots {
		prev, ok := index[s.Slot]
		if !ok || prev.LastFlow != s.LastFlow || prev.Enabled != s.Enabled {
			changed = append(changed, s.Slot)
		}
	}
	return changed
}