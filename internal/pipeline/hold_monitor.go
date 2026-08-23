package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

// HoldBreach records a segment that violated pressure hold limits.
type HoldBreach struct {
	Segment SegmentID
	Bar     float64
	Limit   float64
}

// HoldMonitor tracks pressure hold windows across jumper segments.
type HoldMonitor struct {
	mu       sync.Mutex
	clk      clock.Clock
	hold     *HoldBarontroller
	segments *SegmentRegistry
	sensors  *SensorBank
	breaches []HoldBreach
	windows  []Window
}

func NewHoldMonitor(clk clock.Clock, hold *HoldBarontroller, segments *SegmentRegistry, sensors *SensorBank) *HoldMonitor {
	return &HoldMonitor{clk: clk, hold: hold, segments: segments, sensors: sensors}
}

// BeginHold arms a new pressure hold window and clears prior breach records.
func (m *HoldMonitor) BeginHold(w Window) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hold.Arm(w)
	m.windows = append(m.windows, w)
	m.breaches = nil
}

// Tick evaluates all segments against the active hold window.
func (m *HoldMonitor) Tick(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return model.Wrap("hold_monitor", "canceled", context.Cause(ctx))
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hold.Active() {
		return nil
	}
	id, peak := m.segments.MaxPressure()
	if id == "" {
		return nil
	}
	seg, ok := m.segments.Get(id)
	if !ok {
		return nil
	}
	if len(m.windows) == 0 {
		return nil
	}
	active := m.windows[len(m.windows)-1]
	reading := model.PressureReading{BarGauge: peak, At: m.clk.Now()}
	if active.Contains(reading) && !active.WithinHold(reading) {
		m.breaches = append(m.breaches, HoldBreach{
			Segment: id, Bar: peak, Limit: active.HoldBar,
		})
		return model.Wrap("hold_monitor", string(id),
			fmt.Errorf("hold breach %.1f vs %.1f bar", peak, active.HoldBar))
	}
	if seg.OverPressure() {
		return model.Wrap("hold_monitor", string(id), fmt.Errorf("segment over rated"))
	}
	return nil
}

// Breaches returns a copy of recorded hold violations.
func (m *HoldMonitor) Breaches() []HoldBreach {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HoldBreach, len(m.breaches))
	copy(out, m.breaches)
	return out
}

// ReleaseIfExpired disarms the hold when the window has elapsed.
func (m *HoldMonitor) ReleaseIfExpired() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.hold.Active() {
		m.hold.Release()
		return true
	}
	return false
}

// ActiveWindow returns the current hold window if one is armed.
func (m *HoldMonitor) ActiveWindow() (Window, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.windows) == 0 {
		return Window{}, false
	}
	w := m.windows[len(m.windows)-1]
	return w, m.hold.Active()
}

// Remaining reports time left in the active hold window.
func (m *HoldMonitor) Remaining() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.windows) == 0 {
		return 0
	}
	w := m.windows[len(m.windows)-1]
	return w.Remaining(m.clk)
}
