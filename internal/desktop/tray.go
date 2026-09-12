// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package desktop

import (
	"context"

	"log/slog"
)

// TrayManager handles the application system tray menu and notifications.
type TrayManager struct {
	app *App
}

// NewTrayManager constructs a new TrayManager.
func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{
		app: app,
	}
}

// Start launches system tray integration.
func (t *TrayManager) Start(ctx context.Context) error {
	slog.Info("System Tray Manager initialized")
	return nil
}

// Stop cleanly terminates system tray resources.
func (t *TrayManager) Stop() error {
	slog.Info("System Tray Manager stopped")
	return nil
}
