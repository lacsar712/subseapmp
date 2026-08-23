package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	err = a.ReportLinePressureFault(context.Background(), 8.0)
	if err == nil {
		t.Fatal("expected line pressure high error")
	}
	if !errors.Is(err, model.ErrLinePressureHigh) {
		t.Fatalf("expected ErrLinePressureHigh, got %v", err)
	}
}
