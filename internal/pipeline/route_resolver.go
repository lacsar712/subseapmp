package pipeline

import (
	"fmt"

	"github.com/lacsar712/subseapmp/internal/manifold"
	"github.com/lacsar712/subseapmp/internal/model"
)

// ResolvedRoute is a validated path through jumper segments and manifold valves.
type ResolvedRoute struct {
	From     model.ManifoldID
	To       model.ManifoldID
	Segments []SegmentID
	Valves   []model.ValveID
	Priority int
}

func (r ResolvedRoute) String() string {
	return fmt.Sprintf("%s->%s segs=%d valves=%d", r.From, r.To, len(r.Segments), len(r.Valves))
}

// RouteResolver combines segment topology with manifold valve routing.
type RouteResolver struct {
	segments *SegmentRegistry
	router   *manifold.Router
	planner  *PathPlanner
}

func NewRouteResolver(segments *SegmentRegistry, router *manifold.Router) *RouteResolver {
	return &RouteResolver{
		segments: segments,
		router:   router,
		planner:  NewPathPlanner(segments),
	}
}

// Resolve finds a viable route from source to destination manifold.
func (r *RouteResolver) Resolve(from, to model.ManifoldID) (ResolvedRoute, error) {
	if from == to {
		return ResolvedRoute{From: from, To: to}, nil
	}
	if !r.planner.Reachable(from, to) {
		return ResolvedRoute{}, model.Wrap("route", "unreachable", model.ErrNotFound)
	}
	path, ok := r.router.Path(from, to)
	if !ok {
		return ResolvedRoute{}, model.Wrap("route", "no_valve_path", model.ErrNotFound)
	}
	valves := r.router.ValvesFor(path)
	segs := r.collectSegments(from, to)
	priority := 0
	if len(path) > 0 {
		priority = path[0].Priority
	}
	return ResolvedRoute{
		From: from, To: to, Segments: segs, Valves: valves, Priority: priority,
	}, nil
}

func (r *RouteResolver) collectSegments(from, to model.ManifoldID) []SegmentID {
	var ids []SegmentID
	for _, seg := range r.segments.ForManifold(from) {
		if seg.Blocked {
			continue
		}
		if seg.Upstream == from && seg.Downstream == to {
			ids = append(ids, seg.ID)
		}
		if seg.Downstream == from && seg.Upstream == to {
			ids = append(ids, seg.ID)
		}
	}
	return ids
}

// SegmentPressureCheck verifies no segment on the route exceeds its rated pressure.
func (r *RouteResolver) SegmentPressureCheck(route ResolvedRoute) error {
	for _, segID := range route.Segments {
		seg, ok := r.segments.Get(segID)
		if !ok {
			return model.Wrap("route", string(segID), model.ErrNotFound)
		}
		if seg.OverPressure() {
			return model.Wrap("route", string(segID), fmt.Errorf("over pressure %.1f/%.1f", seg.LastPressure, seg.RatedBar))
		}
		if seg.Blocked {
			return model.Wrap("route", string(segID), model.ErrConflict)
		}
	}
	return nil
}

// Alternatives returns all reachable manifolds from a source via open segments.
func (r *RouteResolver) Alternatives(from model.ManifoldID) []model.ManifoldID {
	seen := map[model.ManifoldID]struct{}{}
	for _, seg := range r.segments.ForManifold(from) {
		if seg.Blocked {
			continue
		}
		if seg.Upstream == from {
			seen[seg.Downstream] = struct{}{}
		}
		if seg.Downstream == from {
			seen[seg.Upstream] = struct{}{}
		}
	}
	out := make([]model.ManifoldID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

// BestRoute picks the highest-priority reachable route among alternatives.
func (r *RouteResolver) BestRoute(from model.ManifoldID) (ResolvedRoute, error) {
	alts := r.Alternatives(from)
	if len(alts) == 0 {
		return ResolvedRoute{}, model.Wrap("route", "no_alternatives", model.ErrNotFound)
	}
	var best ResolvedRoute
	var found bool
	for _, to := range alts {
		route, err := r.Resolve(from, to)
		if err != nil {
			continue
		}
		if err := r.SegmentPressureCheck(route); err != nil {
			continue
		}
		if !found || route.Priority > best.Priority {
			best = route
			found = true
		}
	}
	if !found {
		return ResolvedRoute{}, model.Wrap("route", "no_viable", model.ErrNotFound)
	}
	return best, nil
}
