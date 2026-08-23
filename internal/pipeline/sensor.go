package pipeline

import (
	"sync"

	"github.com/lacsar712/subseapmp/internal/clock"
	"github.com/lacsar712/subseapmp/internal/model"
)

type SensorBank struct {
	mu   sync.RWMutex
	clk  clock.Clock
	data map[model.SensorID]float64
}

func NewSensorBank(clk clock.Clock) *SensorBank {
	return &SensorBank{clk: clk, data: make(map[model.SensorID]float64)}
}

func (b *SensorBank) Set(id model.SensorID, barGauge float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[id] = barGauge
}

func (b *SensorBank) Reading(id model.SensorID) (model.PressureReading, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.data[id]
	if !ok {
		return model.PressureReading{}, false
	}
	return model.PressureReading{Sensor: id, BarGauge: v, At: b.clk.Now()}, true
}

func (b *SensorBank) Average(ids []model.SensorID) (float64, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var sum float64
	var n int
	for _, id := range ids {
		if v, ok := b.data[id]; ok {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}