// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unigodesktop/internal/config"
)

func TestNewApp(t *testing.T) {
	cfg := config.GetDefaultConfig()
	app := NewApp(cfg)

	if app == nil {
		t.Fatalf("expected non-nil App instance")
	}

	if app.GetState() != StateStarting {
		t.Errorf("expected state %s, got %s", StateStarting, app.GetState())
	}
}

func TestAppStartStop(t *testing.T) {
	app := NewApp(nil)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately so Start returns cleanly
	cancel()

	err := app.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error starting/stopping app: %v", err)
	}

	if app.GetState() != StateStopped {
		t.Errorf("expected state %s after stop, got %s", StateStopped, app.GetState())
	}
}
