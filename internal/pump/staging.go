package pump

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

// StagingStep records one booster transition during a staged start sequence.
type StagingStep struct {
	Booster model.BoosterID
	From    model.BoosterState
	To      model.BoosterState
}

// StagingSequence orchestrates multi-booster staging with pressure prerequisites.
type StagingSequence struct {
	mu       sync.Mutex
	clk      clock.Clock
	coord    *BoosterCoordinator
	pressure *BoosterPressureBank
	order    []model.BoosterID
	steps    []StagingStep
}

func NewStagingSequence(clk clock.Clock, coord *BoosterCoordinator, bank *BoosterPressureBank) *StagingSequence {
	return &StagingSequence{clk: clk, coord: coord, pressure: bank}
}

// SetOrder defines the booster start order for staged ramp-up.
func (s *StagingSequence) SetOrder(ids ...model.BoosterID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.order = append([]model.BoosterID(nil), ids...)
}

// StageBatch runs the staging sequence, validating suction before each booster.
func (s *StagingSequence) StageBatch(ctx context.Context) error {
	s.mu.Lock()
	order := append([]model.BoosterID(nil), s.order...)
	s.mu.Unlock()
	if len(order) == 0 {
		return model.Wrap("staging", "empty_order", model.ErrNotFound)
	}
	for _, id := range order {
		if err := s.validatePrerequisites(ctx, id); err != nil {
			return err
		}
		prev := s.coord.States()[id]
		if err := s.coord.Start(ctx, id); err != nil {
			return model.Wrap("staging", string(id), err)
		}
		next := s.coord.States()[id]
		s.mu.Lock()
		s.steps = append(s.steps, StagingStep{Booster: id, From: prev, To: next})
		s.mu.Unlock()
	}
	return nil
}

func (s *StagingSequence) validatePrerequisites(ctx context.Context, id model.BoosterID) error {
	ctrl, ok := s.pressure.Get(id)
	if !ok {
		return nil
	}
	if err := ctrl.Validate(ctx); err != nil {
		return model.Wrap("staging", "pressure", err)
	}
	if !ctrl.WithinSetpoint() {
		return model.Wrap("staging", string(id), fmt.Errorf("pressure out of band"))
	}
	return nil
}

// Steps returns a copy of recorded staging transitions.
func (s *StagingSequence) Steps() []StagingStep {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]StagingStep, len(s.steps))
	copy(out, s.steps)
	return out
}

// Rollback stops boosters in reverse staging order.
func (s *StagingSequence) Rollback(ctx context.Context) error {
	s.mu.Lock()
	order := append([]model.BoosterID(nil), s.order...)
	s.mu.Unlock()
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		state := s.coord.States()[id]
		if state == model.BoosterOff || state == model.BoosterCoast {
			continue
		}
		if err := s.coord.Stop(ctx, id); err != nil {
			return model.Wrap("staging", "rollback", err)
		}
	}
	return nil
}

// ReadyCount returns boosters that reached run state.
func (s *StagingSequence) ReadyCount() int {
	states := s.coord.States()
	n := 0
	for _, id := range s.order {
		if states[id] == model.BoosterRun {
			n++
		}
	}
	return n
}

// AllStaged reports whether every ordered booster is running.
func (s *StagingSequence) AllStaged() bool {
	return s.ReadyCount() == len(s.order) && len(s.order) > 0
}
