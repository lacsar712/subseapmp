package interlock

import (
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/subseapmp/internal/model"
)

type ManifoldLease struct {
	segment  string
	holder   string
	registry *ManifoldLeaseRegistry
	released bool
}

func (l *ManifoldLease) Release() {
	if l.registry == nil || l.released {
		return
	}
	l.registry.release(l.segment, l.holder)
	l.released = true
}

type ManifoldLeaseRegistry struct {
	mu      sync.Mutex
	leases  map[string]string
	expires map[string]time.Time
	ttl     time.Duration
	now     func() time.Time
}

func NewManifoldLeaseRegistry(ttl time.Duration) *ManifoldLeaseRegistry {
	return &ManifoldLeaseRegistry{
		leases: make(map[string]string), expires: make(map[string]time.Time), ttl: ttl,
		now: time.Now,
	}
}

func (r *ManifoldLeaseRegistry) Acquire(segment, holder string, now time.Time) (*ManifoldLease, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[segment]; ok {
		if exp, ok := r.expires[segment]; ok && now.Before(exp) {
			return nil, fmt.Errorf("%s held by %s: %w", segment, h, model.ErrInterlock)
		}
	}
	r.leases[segment] = holder
	r.expires[segment] = now.Add(r.ttl)
	return &ManifoldLease{segment: segment, holder: holder, registry: r}, nil
}

func (r *ManifoldLeaseRegistry) release(segment, holder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[segment]; !ok || h != holder {
		return
	}
	delete(r.leases, segment)
	delete(r.expires, segment)
}

func (r *ManifoldLeaseRegistry) IsHeld(segment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.leases[segment]
	if !ok {
		return false
	}
	exp, ok := r.expires[segment]
	return ok && r.now().Before(exp)
}
