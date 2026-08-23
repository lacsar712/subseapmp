package fsm

import (
	"fmt"

	"github.com/lacsar712/subseapmp/internal/model"
)

type Transition struct {
	From  model.StationState
	To    model.StationState
	Event string
}

var stationTransitions = []Transition{
	{model.StationIdle, model.StationPriming, "prime"},
	{model.StationPriming, model.StationBoosting, "flow_ok"},
	{model.StationBoosting, model.StationHold, "pressure_hold"},
	{model.StationHold, model.StationBoosting, "release_hold"},
	{model.StationBoosting, model.StationIdle, "stop"},
	{model.StationPriming, model.StationFault, "fault"},
	{model.StationBoosting, model.StationFault, "fault"},
	{model.StationHold, model.StationFault, "fault"},
	{model.StationFault, model.StationShutdown, "shutdown"},
	{model.StationIdle, model.StationShutdown, "shutdown"},
}

func AllowedStation(from model.StationState, event string) (model.StationState, bool) {
	for _, t := range stationTransitions {
		if t.From == from && t.Event == event {
			return t.To, true
		}
	}
	return from, false
}

func MustStation(from model.StationState, event string) (model.StationState, error) {
	to, ok := AllowedStation(from, event)
	if !ok {
		return from, model.Wrap("station_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var boosterTransitions = []struct {
	from, to model.BoosterState
	event    string
}{
	{model.BoosterOff, model.BoosterStaging, "start"},
	{model.BoosterStaging, model.BoosterRun, "staged"},
	{model.BoosterRun, model.BoosterCoast, "stop"},
	{model.BoosterCoast, model.BoosterOff, "coast_done"},
	{model.BoosterRun, model.BoosterTrip, "trip"},
	{model.BoosterStaging, model.BoosterTrip, "trip"},
}

func AllowedBooster(from model.BoosterState, event string) (model.BoosterState, bool) {
	for _, t := range boosterTransitions {
		if t.from == from && t.event == event {
			return t.to, true
		}
	}
	return from, false
}