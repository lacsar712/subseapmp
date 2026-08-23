package pipeline

import (
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/model"
)

// SegmentID identifies a riser or jumper segment on the subsea tree.
type SegmentID string

func (id SegmentID) String() string { return string(id) }

// Segment models a pressurized line between manifold headers and the booster intake.
type Segment struct {
	ID           SegmentID
	Upstream     model.ManifoldID
	Downstream   model.ManifoldID
	LengthM      float64
	DiameterMM   float64
	RatedBar     float64
	LastPressure float64
	Blocked      bool
}

func NewSegment(id SegmentID, up, down model.ManifoldID, lengthM, diaMM, ratedBar float64) (*Segment, error) {
	if id == "" || up == "" || down == "" {
		return nil, fmt.Errorf("pipeline: invalid segment identity")
	}
	if lengthM <= 0 || diaMM <= 0 || ratedBar <= 0 {
		return nil, fmt.Errorf("pipeline: segment geometry invalid")
	}
	return &Segment{
		ID: id, Upstream: up, Downstream: down,
		LengthM: lengthM, DiameterMM: diaMM, RatedBar: ratedBar,
	}, nil
}

func (s *Segment) ObservePressure(bar float64) {
	s.LastPressure = bar
}

func (s *Segment) OverPressure() bool {
	return s.LastPressure > s.RatedBar
}

func (s *Segment) Block()  { s.Blocked = true }
func (s *Segment) Unblock() { s.Blocked = false }

// SegmentRegistry tracks all jumper segments tied to a booster station.
type SegmentRegistry struct {
	mu   sync.RWMutex
	byID map[SegmentID]*Segment
}

func NewSegmentRegistry() *SegmentRegistry {
	return &SegmentRegistry{byID: make(map[SegmentID]*Segment)}
}

func (r *SegmentRegistry) Register(seg *Segment) error {
	if seg == nil {
		return fmt.Errorf("pipeline: nil segment")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[seg.ID]; exists {
		return fmt.Errorf("pipeline: duplicate segment %s", seg.ID)
	}
	r.byID[seg.ID] = seg
	return nil
}

func (r *SegmentRegistry) Get(id SegmentID) (*Segment, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byID[id]
	return s, ok
}

func (r *SegmentRegistry) ForManifold(mf model.ManifoldID) []*Segment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Segment
	for _, s := range r.byID {
		if s.Upstream == mf || s.Downstream == mf {
			out = append(out, s)
		}
	}
	return out
}

func (r *SegmentRegistry) BlockedBetween(up, down model.ManifoldID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.byID {
		if s.Upstream == up && s.Downstream == down && s.Blocked {
			return true
		}
	}
	return false
}

func (r *SegmentRegistry) MaxPressure() (SegmentID, float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var bestID SegmentID
	var best float64
	for id, s := range r.byID {
		if s.LastPressure >= best {
			best = s.LastPressure
			bestID = id
		}
	}
	return bestID, best
}

// PathPlanner selects a viable route avoiding blocked segments.
type PathPlanner struct {
	reg *SegmentRegistry
}

func NewPathPlanner(reg *SegmentRegistry) *PathPlanner {
	return &PathPlanner{reg: reg}
}

func (p *PathPlanner) Reachable(from, to model.ManifoldID) bool {
	if from == to {
		return true
	}
	if p.reg.BlockedBetween(from, to) {
		return false
	}
	visited := map[model.ManifoldID]bool{from: true}
	queue := []model.ManifoldID{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, seg := range p.reg.ForManifold(cur) {
			if seg.Blocked {
				continue
			}
			var next model.ManifoldID
			switch cur {
			case seg.Upstream:
				next = seg.Downstream
			case seg.Downstream:
				next = seg.Upstream
			default:
				continue
			}
			if next == to {
				return true
			}
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}
