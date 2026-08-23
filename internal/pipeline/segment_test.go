package pipeline

import (
	"testing"

	"github.com/lacsar712/subseapmp/internal/model"
)

func TestSegmentOverPressure(t *testing.T) {
	seg, err := NewSegment("j1", "mf-a", "mf-b", 120, 150, 180)
	if err != nil {
		t.Fatal(err)
	}
	seg.ObservePressure(200)
	if !seg.OverPressure() {
		t.Fatal("expected over pressure")
	}
}

func TestPathPlannerReachable(t *testing.T) {
	reg := NewSegmentRegistry()
	s1, _ := NewSegment("j1", "mf-a", "mf-b", 100, 120, 200)
	s2, _ := NewSegment("j2", "mf-b", "mf-c", 80, 120, 200)
	_ = reg.Register(s1)
	_ = reg.Register(s2)
	p := NewPathPlanner(reg)
	if !p.Reachable("mf-a", "mf-c") {
		t.Fatal("path should exist")
	}
	s1.Block()
	if p.Reachable("mf-a", "mf-c") {
		t.Fatal("blocked path should fail")
	}
}

func TestSegmentRegistryForManifold(t *testing.T) {
	reg := NewSegmentRegistry()
	s, _ := NewSegment("j1", "mf-x", "mf-y", 50, 100, 150)
	_ = reg.Register(s)
	found := reg.ForManifold(model.ManifoldID("mf-x"))
	if len(found) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(found))
	}
}
