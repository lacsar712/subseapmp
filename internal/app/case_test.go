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
	err = a.HandleManifoldLeak(context.Background(), 12.0)
	if err == nil {
		t.Fatal("expected manifold leak error")
	}
	if !errors.Is(err, model.ErrManifoldLeak) {
		t.Fatalf("expected ErrManifoldLeak, got %v", err)
	}
}
