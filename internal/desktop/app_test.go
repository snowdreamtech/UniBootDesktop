// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"context"
	"testing"
	"time"

	"github.com/snowdreamtech/unibootdesktop/pkg/config"
)

func TestNewApp(t *testing.T) {
	t.Run("with default config via nil", func(t *testing.T) {
		app := NewApp(nil)
		if app == nil {
			t.Fatalf("expected non-nil App instance")
		}
		if app.config == nil {
			t.Fatalf("expected non-nil app.config")
		}
		if app.GetState() != StateStarting {
			t.Errorf("expected initial state %s, got %s", StateStarting, app.GetState())
		}
		if app.runner == nil {
			t.Errorf("expected non-nil runner")
		}
		if app.tray == nil {
			t.Errorf("expected non-nil tray")
		}
		if app.Uptime() < 0 {
			t.Errorf("expected non-negative uptime, got %v", app.Uptime())
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := config.GetDefaultConfig()
		cfg.Debug = true
		cfg.Mode = "local"

		app := NewApp(cfg)
		if app == nil {
			t.Fatalf("expected non-nil App instance")
		}
		if !app.config.Debug {
			t.Errorf("expected debug to be true")
		}
		if app.config.Mode != "local" {
			t.Errorf("expected mode 'local', got %s", app.config.Mode)
		}
	})
}

func TestAppStartStop(t *testing.T) {
	app := NewApp(nil)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately so Start cleanly terminates services
	cancel()

	err := app.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error starting/stopping app: %v", err)
	}

	if state := app.GetState(); state != StateStopped {
		t.Errorf("expected state %s after stop, got %s", StateStopped, state)
	}
}

func TestAppStopIdempotency(t *testing.T) {
	app := NewApp(nil)

	// Stop before start
	if err := app.Stop(); err != nil {
		t.Errorf("expected nil error stopping starting app, got %v", err)
	}

	// Repeated Stop calls should be safe and idempotent
	for i := 0; i < 3; i++ {
		if err := app.Stop(); err != nil {
			t.Errorf("expected nil error on repeated Stop call %d, got %v", i, err)
		}
	}

	if state := app.GetState(); state != StateStopped {
		t.Errorf("expected state %s, got %s", StateStopped, state)
	}
}

func TestAppUptime(t *testing.T) {
	app := NewApp(nil)
	time.Sleep(10 * time.Millisecond)

	uptime := app.Uptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("expected uptime >= 10ms, got %v", uptime)
	}
}
