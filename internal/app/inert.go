package app

import (
	"context"
	"fmt"
	"time"
)

func (a *App) ConfirmInertHold(ctx context.Context, anchor time.Time) error {
	if err := a.inertWindow.Require(anchor); err != nil {
		return fmt.Errorf("inert schedule: %w", err)
	}
	_ = ctx
	return nil
}

func (a *App) InertReady(anchor time.Time) bool {
	return a.inertWindow.Ready(anchor)
}
