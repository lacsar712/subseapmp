package model

import (
	"errors"
	"testing"
)

func TestParseStationID(t *testing.T) {
	id, err := ParseStationID("rack-a")
	if err != nil || id != "rack-a" {
		t.Fatalf("parse Station: %v %q", err, id)
	}
	_, err = ParseStationID("  ")
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid id")
	}
}

func TestFlowSetpointWithin(t *testing.T) {
	sp := FlowSetpoint{LitersPerMinute: 10, TolerancePct: 10}
	if !sp.Within(10.5) {
		t.Fatal("within tolerance")
	}
	if sp.Within(12) {
		t.Fatal("outside tolerance")
	}
}

func TestScheduleClone(t *testing.T) {
	s := BoostSchedule{ID: "s1", Entries: []BoostScheduleEntry{{ID: "e1"}}}
	c := s.Clone()
	c.Entries[0].ID = "mutated"
	if s.Entries[0].ID == "mutated" {
		t.Fatal("clone should be deep for entries slice")
	}
}

func TestWrapUnwrap(t *testing.T) {
	err := Wrap("op", "code", ErrNotFound)
	var de *DomainError
	if !errors.As(err, &de) {
		t.Fatal("expected domain error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatal("unwrap chain")
	}
}