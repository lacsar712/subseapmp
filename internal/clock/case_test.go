package clock

import (
	"testing"
	"time"
)

func TestCase(t *testing.T) {
	start := time.Unix(0, 0)
	clk := NewProcessClock(start, time.Millisecond)
	w := NewInertWindow(clk, 2*time.Second)
	anchor := clk.Now()
	if w.Ready(anchor) {
		t.Fatal("inert window should not be satisfied before process clock advances")
	}
	time.Sleep(3 * time.Second)
	if w.Ready(anchor) {
		t.Fatal("inert window satisfied on wall clock while process clock frozen")
	}
	clk.Advance(3 * time.Second)
	if !w.Ready(anchor) {
		t.Fatal("inert window should be satisfied after process clock advance")
	}
}
