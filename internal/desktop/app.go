// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
)

// AppState represents the current state of the desktop application.
type AppState string

const (
	StateStarting AppState = "starting"
	StateRunning  AppState = "running"
	StateStopping AppState = "stopping"
	StateStopped  AppState = "stopped"
)

// App defines the core desktop application structure.
type App struct {
	mu        sync.RWMutex
	state     AppState
	config    *config.AppConfig
	runner    *UIRunner
	tray      *TrayManager
	startTime time.Time
}

// NewApp initializes a new desktop application instance.
func NewApp(cfg *config.AppConfig) *App {
	if cfg == nil {
		cfg = config.GetDefaultConfig()
	}
	app := &App{
		state:     StateStarting,
		config:    cfg,
		startTime: time.Now(),
	}
	app.runner = NewUIRunner(app)
	app.tray = NewTrayManager(app)
	return app
}

// Start launches the desktop application and its UI/Tray components.
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	a.state = StateRunning
	a.mu.Unlock()

	logger.Info("Starting UniGoDesktop application...", "mode", "desktop")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown signals (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 2)

	// Start UI Runner (Web Bridge Server)
	go func() {
		if err := a.runner.Start(ctx); err != nil {
			errChan <- fmt.Errorf("UI runner error: %w", err)
		}
	}()

	// Start System Tray Manager
	go func() {
		if err := a.tray.Start(ctx); err != nil {
			errChan <- fmt.Errorf("tray manager error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Context canceled, shutting down desktop app...")
	case sig := <-sigChan:
		logger.Info("Signal received, shutting down desktop app...", "signal", sig.String())
	case err := <-errChan:
		logger.Error("Desktop runtime error encountered", "error", err)
		return err
	}

	return a.Stop()
}

// Stop safely shuts down all desktop app services.
func (a *App) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state == StateStopped || a.state == StateStopping {
		return nil
	}

	a.state = StateStopping
	logger.Info("Stopping UniGoDesktop components...")

	if err := a.runner.Stop(); err != nil {
		logger.Warn("Failed to cleanly stop UI runner", "error", err)
	}

	if err := a.tray.Stop(); err != nil {
		logger.Warn("Failed to cleanly stop tray manager", "error", err)
	}

	a.state = StateStopped
	logger.Info("UniGoDesktop application stopped cleanly")
	return nil
}

// GetState returns the current application state.
func (a *App) GetState() AppState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Uptime returns the duration since application start.
func (a *App) Uptime() time.Duration {
	return time.Since(a.startTime)
}
