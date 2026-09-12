// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"log/slog"

	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/installer"
	"github.com/snowdreamtech/unigodesktop/pkg/qemu"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
)

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails application starts up.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("UniGoDesktop Wails GUI runtime started successfully")
}

// GetDiskList returns all removable USB drives safely filtered.
func (a *App) GetDiskList() ([]disk.DiskInfo, error) {
	return disk.GetRemovableDisks()
}

// DeployModeA triggers Mode A (Hybrid Pro Mode - Ventoy + iPXE).
func (a *App) DeployModeA(targetDisk string) (*installer.DeployResult, error) {
	return installer.DeployModeA(a.ctx, targetDisk)
}

// DeployModeB triggers Mode B (Cloud Pure Mode - 1-sec FAT32 + iPXE).
func (a *App) DeployModeB(targetDisk string) (*installer.DeployResult, error) {
	return installer.DeployModeB(a.ctx, targetDisk)
}

// CheckQEMU returns QEMU detection metadata.
func (a *App) CheckQEMU() *qemu.QEMUStatus {
	return qemu.Detect()
}

// CheckUpdate returns GitHub release update metadata.
func (a *App) CheckUpdate() *updater.UpdateStatus {
	return updater.CheckUpdate(a.ctx)
}

// GetConfig loads the application settings.
func (a *App) GetConfig() (*config.AppConfig, error) {
	return config.Load()
}
