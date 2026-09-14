// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"log/slog"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/firmware"
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

// SelectIsoFiles opens a native multi-file open dialog for selecting Ventoy-supported system image files (.iso, .wim, .img, .vhd, etc.).
func (a *App) SelectIsoFiles() ([]string, error) {
	return wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select System Image Files (*.iso, *.wim, *.img, *.vhd, etc.)",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Ventoy Source Images (*.iso; *.wim; *.img; *.vhd; *.vhdx; *.vti; *.efi; *.bin; *.xz; *.gz; *.raw)",
				Pattern:     "*.iso;*.wim;*.img;*.vhd;*.vhdx;*.vti;*.efi;*.bin;*.xz;*.gz;*.raw",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
}

// DeployModeA triggers Mode A (Hybrid Pro Mode - Ventoy + iPXE) with customizable file system and optional ISO files.
func (a *App) DeployModeA(targetDisk string, fsType string, isoPaths []string) (*installer.DeployResult, error) {
	cfg, _ := config.Load()
	ventoyPath := ""
	if cfg != nil {
		ventoyPath = cfg.VentoyPath
	}
	progressCb := func(p installer.IsoCopyProgress) {
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	return installer.DeployModeAWithIsoAndVentoyPath(a.ctx, targetDisk, fsType, ventoyPath, isoPaths, progressCb)
}

// ValidateVentoyCli verifies the user-specified Ventoy CLI path.
func (a *App) ValidateVentoyCli(ventoyPath string) *installer.VentoyCliValidationResult {
	return installer.ValidateVentoyCli(ventoyPath)
}

// DeployModeABatch triggers Mode A deployment for multiple target USB drives with customizable file system and optional ISO files.
func (a *App) DeployModeABatch(targetDisks []string, fsType string, isoPaths []string) ([]*installer.DeployResult, error) {
	cfg, _ := config.Load()
	ventoyPath := ""
	if cfg != nil {
		ventoyPath = cfg.VentoyPath
	}
	progressCb := func(p installer.IsoCopyProgress) {
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	return installer.DeployModeABatchWithVentoyAndIso(a.ctx, targetDisks, fsType, ventoyPath, isoPaths, progressCb)
}

// DeployModeB triggers Mode B (Cloud Pure Mode) with customizable file system.
func (a *App) DeployModeB(targetDisk string, fsType string) (*installer.DeployResult, error) {
	return installer.DeployModeB(a.ctx, targetDisk, fsType)
}

// DeployModeBBatch triggers Mode B deployment for multiple target USB drives with customizable file system.
func (a *App) DeployModeBBatch(targetDisks []string, fsType string) ([]*installer.DeployResult, error) {
	return installer.DeployModeBBatch(a.ctx, targetDisks, fsType)
}

// CheckQEMU returns QEMU detection metadata.
func (a *App) CheckQEMU() *qemu.QEMUStatus {
	return qemu.Detect()
}

// LaunchQEMU triggers a QEMU virtual machine test instance for the target USB drive.
func (a *App) LaunchQEMU(targetDisk string) error {
	return qemu.LaunchTest(a.ctx, targetDisk)
}

// CheckUpdate returns GitHub release update metadata.
func (a *App) CheckUpdate() *updater.UpdateStatus {
	return updater.CheckUpdate(a.ctx)
}

// GetConfig loads the application settings.
func (a *App) GetConfig() (*config.AppConfig, error) {
	return config.Load()
}

// SaveConfig updates and saves application settings.
func (a *App) SaveConfig(cfg *config.AppConfig) error {
	if cfg == nil {
		return config.GetDefaultConfig().Save()
	}
	return cfg.Save()
}

// GetFirmwareList returns the standard UniBoot firmware mapping matrix.
func (a *App) GetFirmwareList() []firmware.FirmwareMapping {
	return firmware.GetFirmwareMappings()
}

// GetUniBootReleaseInfo queries the latest UniBoot GitHub release metadata.
func (a *App) GetUniBootReleaseInfo() (*firmware.UniBootReleaseInfo, error) {
	cfg, _ := config.Load()
	proxy := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
	}
	return firmware.FetchLatestUniBootRelease(a.ctx, proxy)
}

// SyncUniBootFirmware downloads the latest UniBoot release firmware assets to local cache.
func (a *App) SyncUniBootFirmware() (*firmware.UniBootReleaseInfo, error) {
	cfg, _ := config.Load()
	proxy := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
	}
	return firmware.SyncUniBootFirmware(a.ctx, proxy)
}

