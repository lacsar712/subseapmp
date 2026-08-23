package manifold

import (
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/model"
)

// SlotCoordinator aligns manifold slot assignments with production headers.
type SlotCoordinator struct {
	mu      sync.Mutex
	slots   *SlotTable
	headers *HeaderTable
	flow    map[model.SlotID]model.FlowSetpoint
}

func NewSlotCoordinator(slots *SlotTable, headers *HeaderTable) *SlotCoordinator {
	return &SlotCoordinator{
		slots:   slots,
		headers: headers,
		flow:    make(map[model.SlotID]model.FlowSetpoint),
	}
}

// Assign binds a slot index to a manifold and records the flow setpoint.
func (c *SlotCoordinator) Assign(index int, mf model.ManifoldID, sp model.FlowSetpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.slots.Assign(index, mf); err != nil {
		return err
	}
	slot := c.slots.Slots()[index]
	c.flow[slot.ID] = sp
	return nil
}

// Enable toggles a slot and releases its header when disabled.
func (c *SlotCoordinator) Enable(index int, on bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.slots.Enable(index, on); err != nil {
		return err
	}
	if !on {
		slot := c.slots.Slots()[index]
		for _, h := range c.headers.ByManifold(slot.Manifold) {
			if h.Allocated {
				h.Release()
			}
		}
	}
	return nil
}

// SyncHeaders allocates production headers for every enabled slot.
func (c *SlotCoordinator) SyncHeaders() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, slot := range c.slots.Slots() {
		if !slot.Enabled {
			continue
		}
		sp, ok := c.flow[slot.ID]
		if !ok {
			return model.Wrap("slot_coord", string(slot.ID), model.ErrFlowSetpoint)
		}
		hdr, err := c.headers.PickAvailable(slot.Manifold)
		if err != nil {
			return model.Wrap("slot_coord", "header", err)
		}
		if err := hdr.Allocate(sp.LitersPerMinute); err != nil {
			return model.Wrap("slot_coord", string(hdr.ID), err)
		}
	}
	return nil
}

// CoordinatedFlow returns the total flow allocated across enabled slots.
func (c *SlotCoordinator) CoordinatedFlow() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var total float64
	for _, slot := range c.slots.Slots() {
		if !slot.Enabled {
			continue
		}
		if sp, ok := c.flow[slot.ID]; ok {
			total += sp.LitersPerMinute
		}
	}
	return total
}

// ValidateSlotAlignment checks that enabled slots have matching header allocations.
func (c *SlotCoordinator) ValidateSlotAlignment() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, slot := range c.slots.Slots() {
		if !slot.Enabled {
			continue
		}
		headers := c.headers.ByManifold(slot.Manifold)
		allocated := 0
		for _, h := range headers {
			if h.Allocated {
				allocated++
			}
		}
		if allocated == 0 {
			return model.Wrap("slot_coord", string(slot.ID), fmt.Errorf("no header for enabled slot"))
		}
	}
	return nil
}

// SlotAssignments builds model assignments for snapshot persistence.
func (c *SlotCoordinator) SlotAssignments() []model.SlotAssignment {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []model.SlotAssignment
	for _, slot := range c.slots.Slots() {
		sp := c.flow[slot.ID]
		out = append(out, model.SlotAssignment{
			Slot: slot.ID, Manifold: slot.Manifold, Enabled: slot.Enabled, Setpoint: sp,
		})
	}
	return out
}
