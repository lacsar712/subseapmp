package fsm

import (
	"context"

	"github.com/lacsar712/subseapmp/internal/model"
)

type StationSideEffect func(ctx context.Context, rack model.StationID, from, to model.StationState) error

type StationFSM struct {
	id       model.StationID
	state    model.StationState
	onChange StationSideEffect
}

func NewStationFSM(id model.StationID, effect StationSideEffect) *StationFSM {
	return &StationFSM{id: id, state: model.StationIdle, onChange: effect}
}

func (f *StationFSM) State() model.StationState { return f.state }

func (f *StationFSM) Apply(ctx context.Context, event string) error {
	next, err := MustStation(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("station_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}