package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/lacsar712/subseapmp/internal/config"
	"github.com/lacsar712/subseapmp/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	fails := 0
	CalibrateProbe = func(ctx context.Context) error {
		fails++
		if fails == 1 {
			return fmt.Errorf("sensor fault")
		}
		return nil
	}
	defer func() { CalibrateProbe = nil }()
	ctx := context.Background()
	segment := model.ManifoldID("mf-primary")
	if err := a.CalibrateManifold(ctx, segment, "crew-a"); err == nil {
		t.Fatal("expected calibration failure")
	}
	if err := a.CalibrateManifold(ctx, segment, "crew-b"); err != nil {
		t.Fatalf("second holder blocked by leaked manifold lease: %v", err)
	}
}
