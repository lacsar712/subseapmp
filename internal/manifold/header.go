package manifold

import (
	"fmt"
	"sync"

	"github.com/lacsar712/subseapmp/internal/model"
)

// HeaderID names a production header on the subsea manifold tree.
type HeaderID string

func (h HeaderID) String() string { return string(h) }

// Header groups valve positions feeding a booster intake line.
type Header struct {
	ID        HeaderID
	Manifold  model.ManifoldID
	Valve     model.ValveID
	Priority  int
	Allocated bool
	FlowLPM   float64
}

func NewHeader(id HeaderID, mf model.ManifoldID, valve model.ValveID, priority int) (*Header, error) {
	if id == "" || mf == "" || valve == "" {
		return nil, fmt.Errorf("manifold: invalid header")
	}
	return &Header{ID: id, Manifold: mf, Valve: valve, Priority: priority}, nil
}

func (h *Header) Allocate(flow float64) error {
	if h.Allocated {
		return model.ErrConflict
	}
	h.Allocated = true
	h.FlowLPM = flow
	return nil
}

func (h *Header) Release() {
	h.Allocated = false
	h.FlowLPM = 0
}

// HeaderTable manages production headers for a booster station manifold.
type HeaderTable struct {
	mu      sync.RWMutex
	headers []*Header
}

func NewHeaderTable(items ...*Header) *HeaderTable {
	t := &HeaderTable{}
	for _, h := range items {
		t.headers = append(t.headers, h)
	}
	return t
}

func (t *HeaderTable) Add(h *Header) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.headers = append(t.headers, h)
}

func (t *HeaderTable) ByManifold(mf model.ManifoldID) []*Header {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []*Header
	for _, h := range t.headers {
		if h.Manifold == mf {
			out = append(out, h)
		}
	}
	return out
}

func (t *HeaderTable) PickAvailable(mf model.ManifoldID) (*Header, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var best *Header
	for _, h := range t.headers {
		if h.Manifold != mf || h.Allocated {
			continue
		}
		if best == nil || h.Priority > best.Priority {
			best = h
		}
	}
	if best == nil {
		return nil, model.ErrNotFound
	}
	return best, nil
}

func (t *HeaderTable) TotalAllocatedFlow() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var sum float64
	for _, h := range t.headers {
		if h.Allocated {
			sum += h.FlowLPM
		}
	}
	return sum
}
