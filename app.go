// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/firmware"
	"github.com/snowdreamtech/unigodesktop/pkg/hypervisor"
	"github.com/snowdreamtech/unigodesktop/pkg/installer"
	"github.com/snowdreamtech/unigodesktop/pkg/qemu"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
	"github.com/snowdreamtech/unigodesktop/pkg/utils"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppInfo contains dynamic build and environment metadata.
type AppInfo struct {
	ProjectName    string `json:"projectName"`
	Version        string `json:"version"`
	GitTag         string `json:"gitTag"`
	CommitHash     string `json:"commitHash"`
	CommitHashFull string `json:"commitHashFull"`
	BuildTime      string `json:"buildTime"`
	Author         string `json:"author"`
	Copyright      string `json:"copyright"`
	License        string `json:"license"`
	GoVersion      string `json:"goVersion"`
	OsArch         string `json:"osArch"`
}

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx              context.Context
	deployCancelFunc context.CancelFunc
	cancelMutex      sync.Mutex
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

func (a *App) initDeployContext() context.Context {
	a.cancelMutex.Lock()
	defer a.cancelMutex.Unlock()
	ctx, cancel := context.WithCancel(a.ctx)
	a.deployCancelFunc = cancel
	return ctx
}

func (a *App) clearDeployContext() {
	a.cancelMutex.Lock()
	defer a.cancelMutex.Unlock()
	a.deployCancelFunc = nil
}

// CancelDeployment cancels any active in-flight deployment task.
func (a *App) CancelDeployment() bool {
	a.cancelMutex.Lock()
	defer a.cancelMutex.Unlock()
	if a.deployCancelFunc != nil {
		logger.Info("User requested active deployment cancellation")
		a.deployCancelFunc()
		a.deployCancelFunc = nil
		return true
	}
	return false
}

// startup is called when the Wails application starts up.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.SetWailsContext(ctx)
	logger.Info(fmt.Sprintf("UniGoDesktop Wails GUI runtime started successfully (%s/%s)", runtime.GOOS, runtime.GOARCH))
	disk.StartHotplugMonitor(ctx, func() {
		logger.Info("Removable disk change detected, refreshing drive list")
		wailsRuntime.EventsEmit(a.ctx, "disk-list-changed")
	})
}

// shutdown is called automatically when the Wails application is closing.
func (a *App) shutdown(ctx context.Context) {
	logger.Info("UniGoDesktop Wails GUI runtime shutting down, performing hypervisor disk cleanup")
	hypervisor.GetManager().CleanupAllUnmountedDisks()
}

// GetRecentLogs returns recent log entries from the memory buffer.
func (a *App) GetRecentLogs() []logger.LogEntry {
	return logger.GetRecentLogs()
}

// ClearLogs clears the in-memory log buffer.
func (a *App) ClearLogs() {
	logger.Info("User cleared in-memory log history")
	logger.ClearLogs()
}

// LogAction allows the frontend to log user UI interaction events directly into the Log Center.
func (a *App) LogAction(level string, message string, details string) {
	if level == "" {
		level = "INFO"
	}
	if details != "" {
		logger.RecordLog(level, message, "details", details)
	} else {
		logger.RecordLog(level, message)
	}
}

// GetDiskList returns all removable disks safely filtered.
func (a *App) GetDiskList() ([]disk.DiskInfo, error) {
	disks, err := disk.GetRemovableDisks()
	if err != nil {
		logger.Error("Failed to scan removable storage drives", "error", err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("Scanned removable storage drives, found %d device(s)", len(disks)))
	return disks, nil
}

// EjectDisk safely unmounts and ejects the target removable storage disk.
func (a *App) EjectDisk(targetDisk string) error {
	logger.Info("Requesting safe ejection for disk", "disk", targetDisk)
	disk.InvalidateDiskCache()
	err := disk.EjectDisk(targetDisk)
	if err != nil {
		logger.Error("Failed to eject target disk", "disk", targetDisk, "error", err)
		return err
	}
	logger.Info("Target disk safely ejected", "disk", targetDisk)
	return nil
}

// SelectIsoFiles opens a native multi-file open dialog for selecting Ventoy-supported system image files (.iso, .wim, .img, .vhd, etc.).
func (a *App) SelectIsoFiles(title string, ventoyFilter string, allFilter string) ([]string, error) {
	if title == "" {
		title = "Select System Image Files (*.iso, *.wim, *.img, *.vhd, etc.)"
	}
	if ventoyFilter == "" {
		ventoyFilter = "Ventoy Source Images (*.iso; *.wim; *.img; *.vhd; *.vhdx; *.vti; *.efi; *.bin; *.xz; *.gz; *.raw)"
	}
	if allFilter == "" {
		allFilter = "All Files (*.*)"
	}

	logger.Info("Opening native system image file picker dialog")
	paths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: ventoyFilter,
				Pattern:     "*.iso;*.wim;*.img;*.vhd;*.vhdx;*.vti;*.efi;*.bin;*.xz;*.gz;*.raw",
			},
			{
				DisplayName: allFilter,
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		logger.Error("Failed to open system image file picker", "error", err)
		return nil, err
	}
	if len(paths) > 0 {
		logger.Info(fmt.Sprintf("Selected %d system image file(s)", len(paths)))
	}
	return paths, nil
}

// CalculateFileChecksum computes MD5, SHA256, or SHA512 hash for the specified image file.
func (a *App) CalculateFileChecksum(filePath string, algo string) (*utils.ChecksumResult, error) {
	logger.Info("Calculating file checksum", "filePath", filePath, "algorithm", algo)
	res, err := utils.CalculateFileChecksum(a.ctx, filePath, algo)
	if err != nil {
		logger.Error("Failed to calculate file checksum", "filePath", filePath, "algorithm", algo, "error", err)
		return nil, err
	}
	logger.Info("File checksum calculated successfully", "filePath", filePath, "algo", res.Algorithm, "hash", res.Hash, "durationMs", res.DurationMs)
	return res, nil
}

// ExportLogs opens a native save file dialog to export log content to a file (.log or .txt).
func (a *App) ExportLogs(content string, title string, logFilter string, textFilter string, allFilter string) (string, error) {
	if title == "" {
		title = "Export Log File"
	}
	if logFilter == "" {
		logFilter = "Log Files (*.log)"
	}
	if textFilter == "" {
		textFilter = "Text Files (*.txt)"
	}
	if allFilter == "" {
		allFilter = "All Files (*.*)"
	}

	defaultFilename := fmt.Sprintf("unigodesktop-log-%s.log", time.Now().Format("2006-01-02-150405"))
	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: logFilter,
				Pattern:     "*.log",
			},
			{
				DisplayName: textFilter,
				Pattern:     "*.txt",
			},
			{
				DisplayName: allFilter,
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		logger.Error("Failed to open save file dialog for log export", "error", err)
		return "", fmt.Errorf("open save file dialog: %w", err)
	}
	if filePath == "" {
		return "", nil // User cancelled
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		logger.Error("Failed to write log export file", "path", filePath, "error", err)
		return "", fmt.Errorf("write log file: %w", err)
	}
	logger.Info("Logs exported successfully", "path", filePath)
	return filePath, nil
}

// DeployHybridMode triggers Hybrid Mode (Hybrid Pro Mode - Ventoy + iPXE) with customizable file system and optional ISO files.
func (a *App) DeployHybridMode(targetDisk string, fsType string, isoPaths []string, expected disk.DiskInfo) (*installer.DeployResult, error) {
	logger.Info("User confirmed Hybrid Mode boot disk creation", "disk", targetDisk, "fs", fsType, "isoCount", len(isoPaths))
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	cfg, _ := config.Load()
	ventoyPath := ""
	if cfg != nil {
		ventoyPath = cfg.VentoyPath
	}
	progressCb := func(p installer.IsoCopyProgress) {
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	res, err := installer.DeployHybridModeWithExpectedDisk(deployCtx, targetDisk, fsType, ventoyPath, isoPaths, progressCb, expected)
	if err != nil && res != nil {
		return res, nil
	}
	return res, err
}

// ValidateVentoyCli verifies the user-specified Ventoy CLI path.
func (a *App) ValidateVentoyCli(ventoyPath string) *installer.VentoyCliValidationResult {
	return installer.ValidateVentoyCli(ventoyPath)
}

// DeployHybridModeBatch triggers Hybrid Mode deployment for multiple target disk drives with customizable file system and optional ISO files.
func (a *App) DeployHybridModeBatch(targetDisks []string, fsType string, isoPaths []string, expected []disk.DiskInfo) ([]*installer.DeployResult, error) {
	logger.Info("User confirmed batch Hybrid Mode boot disk creation", "diskCount", len(targetDisks), "fs", fsType)
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	cfg, _ := config.Load()
	ventoyPath := ""
	if cfg != nil {
		ventoyPath = cfg.VentoyPath
	}
	progressCb := func(p installer.IsoCopyProgress) {
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	return installer.DeployHybridModeBatchWithExpectedDisks(deployCtx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, expected)
}

// DeployCloudMode triggers Cloud Mode (Cloud Pure Mode) with customizable file system.
func (a *App) DeployCloudMode(targetDisk string, fsType string, expected disk.DiskInfo) (*installer.DeployResult, error) {
	logger.Info("User confirmed Cloud Mode boot disk creation", "disk", targetDisk, "fs", fsType)
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	// Emit stage-based progress events so the frontend can show real progress
	// instead of a fake interval timer.
	wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-progress", 10) // started: validating disk
	res, err := installer.DeployCloudModeWithExpectedDisk(deployCtx, targetDisk, fsType, expected)
	if err == nil {
		wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-progress", 100) // complete
	}
	if err != nil && res != nil {
		return res, nil
	}
	return res, err
}

// DeployCloudModeBatch triggers Cloud Mode deployment for multiple target disk drives with customizable file system.
func (a *App) DeployCloudModeBatch(targetDisks []string, fsType string, expected []disk.DiskInfo) ([]*installer.DeployResult, error) {
	logger.Info("User confirmed batch Cloud Mode boot disk creation", "diskCount", len(targetDisks), "fs", fsType)
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-progress", 10) // started
	results, err := installer.DeployCloudModeBatchWithExpectedDisks(deployCtx, targetDisks, fsType, expected)
	if err == nil {
		wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-progress", 100) // complete
	}
	return results, err
}

// CheckQEMU returns QEMU detection metadata for backward compatibility.
func (a *App) CheckQEMU() *qemu.QEMUStatus {
	best := hypervisor.GetManager().DetectBest()
	if best != nil && best.Installed {
		return &qemu.QEMUStatus{
			Installed: true,
			Path:      best.Path,
			Version:   best.Version,
		}
	}
	return qemu.Detect()
}

// DetectHypervisors returns status of all installed virtual machine engines.
func (a *App) DetectHypervisors() []*hypervisor.VMStatus {
	return hypervisor.GetManager().DetectAll()
}

// DetectBestHypervisor returns the highest priority available virtual machine status.
func (a *App) DetectBestHypervisor() *hypervisor.VMStatus {
	return hypervisor.GetManager().DetectBest()
}

// LaunchQEMU triggers virtual machine test instance using highest priority available hypervisor (QEMU, UTM, VMware, VirtualBox).
func (a *App) LaunchQEMU(targetDisk string) error {
	logger.Info("Requesting hypervisor preview test launch", "disk", targetDisk)
	err := hypervisor.GetManager().LaunchBest(a.ctx, targetDisk, hypervisor.BootModeAuto)
	if err != nil {
		logger.Error("Failed to launch hypervisor preview test", "disk", targetDisk, "error", err)
		return err
	}
	logger.Info("Hypervisor preview test launched successfully", "disk", targetDisk)
	return nil
}

// LaunchVM launches a specified or best available virtual machine with boot mode (uefi, bios, auto).
func (a *App) LaunchVM(targetDisk string, vmType string, bootMode string) error {
	if bootMode == "" {
		bootMode = hypervisor.BootModeAuto
	}
	cfg := hypervisor.VMConfig{
		CpuCores:     hypervisor.GetRecommendedVCPUs(),
		MemoryMB:     hypervisor.GetRecommendedVMMemoryMB(),
		BootMode:     bootMode,
		DisplayAccel: true,
		SecureBoot:   false,
	}
	return a.LaunchVMWithConfig(targetDisk, vmType, cfg)
}

// GetDefaultVMConfig returns recommended default VM tuning parameters.
func (a *App) GetDefaultVMConfig() *hypervisor.VMConfig {
	return hypervisor.DefaultVMConfig()
}

// LaunchVMWithConfig launches a virtual machine with full custom VMConfig options.
func (a *App) LaunchVMWithConfig(targetDisk string, vmType string, cfg hypervisor.VMConfig) error {
	if cfg.BootMode == "" {
		cfg.BootMode = hypervisor.BootModeAuto
	}
	if cfg.CpuCores <= 0 {
		cfg.CpuCores = hypervisor.GetRecommendedVCPUs()
	}
	if cfg.MemoryMB <= 0 {
		cfg.MemoryMB = hypervisor.GetRecommendedVMMemoryMB()
	}

	if vmType == "" || vmType == "auto" {
		return hypervisor.GetManager().LaunchBestConfigured(a.ctx, targetDisk, cfg)
	}
	logger.Info("Requesting specified hypervisor preview test launch with VMConfig", "disk", targetDisk, "vmType", vmType, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)
	err := hypervisor.GetManager().LaunchSpecifiedConfigured(a.ctx, targetDisk, hypervisor.HypervisorType(vmType), cfg)
	if err != nil {
		logger.Error("Failed to launch specified hypervisor with VMConfig", "disk", targetDisk, "vmType", vmType, "error", err)
		return err
	}
	logger.Info("Specified hypervisor test launched successfully with VMConfig", "disk", targetDisk, "vmType", vmType)
	return nil
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
		logger.Info("Resetting application settings to default")
		return config.GetDefaultConfig().Save()
	}
	logger.Info("Saving updated application preferences", "language", cfg.Language, "theme", cfg.Theme, "autoEject", cfg.AutoEjectAfterDeploy, "proxy", cfg.GithubProxy)
	if cfg.ProxyPassword != "" {
		if err := config.SaveProxyPassword(cfg.ProxyPassword); err != nil {
			logger.Error("Failed to save proxy password", "error", err)
			return err
		}
		cfg.ProxyPassword = ""
	}
	err := cfg.Save()
	if err != nil {
		logger.Error("Failed to save application config", "error", err)
		return err
	}
	logger.Info("Application preferences saved successfully")
	return nil
}

// ClearProxyPassword removes the saved proxy password from the system credential store.
func (a *App) ClearProxyPassword() error {
	return config.DeleteProxyPassword()
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

// GetAppInfo returns dynamic build, environment, and version metadata.
func (a *App) GetAppInfo() AppInfo {
	versionStr := env.GitTag
	if versionStr == "" {
		versionStr = "N/A"
	}
	gitTag := env.GitTag
	if gitTag == "" {
		gitTag = "N/A"
	}
	commitHash := env.CommitHash
	if commitHash == "" {
		commitHash = "N/A"
	}
	commitHashFull := env.CommitHashFull
	if commitHashFull == "" {
		commitHashFull = "N/A"
	}
	buildTime := env.BuildTime
	if buildTime == "" {
		buildTime = "N/A"
	}

	return AppInfo{
		ProjectName:    env.ProjectName,
		Version:        versionStr,
		GitTag:         gitTag,
		CommitHash:     commitHash,
		CommitHashFull: commitHashFull,
		BuildTime:      buildTime,
		Author:         env.Author,
		Copyright:      env.COPYRIGHT,
		License:        env.LICENSE,
		GoVersion:      runtime.Version(),
		OsArch:         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// PerformGuiUpdate downloads the latest GUI release package with progress events.
func (a *App) PerformGuiUpdate() (*updater.GuiUpdateResult, error) {
	cfg, _ := config.Load()
	proxy := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
	}
	progressCb := func(p updater.GuiUpdateProgress) {
		wailsRuntime.EventsEmit(a.ctx, "gui-update-progress", p)
	}
	return updater.PerformGuiUpdate(a.ctx, proxy, progressCb)
}

// OpenBrowserURL opens the target URL in the user's default system browser.
func (a *App) OpenBrowserURL(targetURL string) {
	wailsRuntime.BrowserOpenURL(a.ctx, targetURL)
}

// OpenAboutModal emits an event to the frontend to trigger the About dialog.
func (a *App) OpenAboutModal() {
	wailsRuntime.EventsEmit(a.ctx, "open-about-modal")
}

// ReloadAppMenu rebuilds and updates the native application menu with the specified language.
func (a *App) ReloadAppMenu(lang string) error {
	if a.ctx == nil {
		return nil
	}
	appMenu := BuildAppMenu(a, lang)
	wailsRuntime.MenuSetApplicationMenu(a.ctx, appMenu)
	wailsRuntime.MenuUpdateApplicationMenu(a.ctx)
	return nil
}
