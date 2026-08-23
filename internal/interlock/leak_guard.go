package interlock

import "github.com/lacsar712/subseapmp/internal/model"

type ManifoldLeakGuard struct {
	threshold float64
}

func NewManifoldLeakGuard(threshold float64) *ManifoldLeakGuard {
	return &ManifoldLeakGuard{threshold: threshold}
}

func (g *ManifoldLeakGuard) Permit(delta float64) error {
	if delta > g.threshold {
		return model.Wrap("manifold_leak", "detected", model.ErrManifoldLeak)
	}
	return nil
}
