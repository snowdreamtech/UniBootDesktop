// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unibootdesktop/internal/env"
	"github.com/snowdreamtech/unibootdesktop/internal/logger"
	"github.com/snowdreamtech/unibootdesktop/pkg/config"
	"github.com/snowdreamtech/unibootdesktop/pkg/disk"
	"github.com/snowdreamtech/unibootdesktop/pkg/firmware"
	"github.com/snowdreamtech/unibootdesktop/pkg/hypervisor"
	"github.com/snowdreamtech/unibootdesktop/pkg/installer"
	"github.com/snowdreamtech/unibootdesktop/pkg/privilege"
	"github.com/snowdreamtech/unibootdesktop/pkg/updater"
	"github.com/snowdreamtech/unibootdesktop/pkg/utils"
	"github.com/wailsapp/wails/v2/pkg/options"
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

// SystemInfo holds runtime and OS metadata.
type SystemInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"goVersion"`
	AppName   string `json:"appName"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	DataDir   string `json:"dataDir"`
	ConfigDir string `json:"configDir"`
}

// NetworkTestResult holds connectivity test metrics.
type NetworkTestResult struct {
	Connected bool   `json:"connected"`
	LatencyMs int64  `json:"latencyMs"`
	TargetURL string `json:"targetUrl"`
	Error     string `json:"error,omitempty"`
}

// HelloInfo provides greeting and runtime demonstration.
type HelloInfo struct {
	Greeting  string `json:"greeting"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Timestamp string `json:"timestamp"`
}

// UniBootStatus represents the lightweight cross-platform ready status of native UniBoot engine.
type UniBootStatus struct {
	Ready   bool   `json:"ready"`
	Version string `json:"version"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// BatchEjectResult describes the outcome of a concurrent multi-disk ejection.
type BatchEjectResult struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"`
}

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx              context.Context
	cancel           context.CancelFunc
	deployCancelFunc context.CancelFunc
	cancelMutex      sync.Mutex
	mu               sync.Mutex
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

func (a *App) emitEvent(eventName string, optionalData ...interface{}) {
	if a.ctx != nil && a.ctx.Value("frontend") != nil {
		wailsRuntime.EventsEmit(a.ctx, eventName, optionalData...)
	}
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
	a.ctx, a.cancel = context.WithCancel(ctx)
	logger.SetWailsContext(ctx)
	logger.Info(fmt.Sprintf("UniBootDesktop Wails GUI runtime started successfully (%s/%s)", runtime.GOOS, runtime.GOARCH))
	if err := env.EnsureAppDirs(); err != nil {
		logger.Warn("Failed to ensure default app directories exist", "error", err)
	}
	if cfg, err := config.Load(); err == nil && cfg != nil && cfg.UniBootPath != "" {
		firmware.SetCustomUniBootDir(cfg.UniBootPath)
	}
	disk.StartHotplugMonitor(ctx, func() {
		logger.Info("Removable disk change detected, refreshing drive list")
		a.emitEvent("disk-list-changed")
	})

	hypervisor.RegisterVMExitHandler(func(targetDisk string, vmErr error) {
		logger.Info("Hypervisor VM session terminated, bouncing UI back and refreshing disk state", "disk", targetDisk)
		disk.InvalidateDiskCache()
		if a.ctx != nil {
			wailsRuntime.WindowUnminimise(a.ctx)
			wailsRuntime.WindowShow(a.ctx)
			a.emitEvent("disk-list-changed")
			a.emitEvent("vm-session-ended", map[string]interface{}{
				"disk":    targetDisk,
				"success": vmErr == nil,
				"error": func() string {
					if vmErr != nil {
						return vmErr.Error()
					}
					return ""
				}(),
			})
		}
	})
}

// shutdown is called automatically when the Wails application is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	logger.Info("UniBootDesktop Wails GUI runtime shutting down")

	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		hypervisor.GetManager().CleanupAllUnmountedDisks()
	}()

	select {
	case <-cleanupDone:
		logger.Info("Tracked hypervisor disk cleanup completed before shutdown timeout")
	case <-time.After(5 * time.Second):
		logger.Warn("Tracked hypervisor disk cleanup hit shutdown timeout; continuing app exit")
	}

	if client := privilege.GetActiveWorkerClient(); client != nil {
		_ = client.Close()
		privilege.SetActiveWorkerClient(nil)
	}
}

// beforeClose is invoked before the application window closes.
// If EnableTray is true and CloseAction is "minimize_to_tray", the window is hidden instead of exiting.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	if cfg.EnableTray && cfg.CloseAction == "minimize_to_tray" {
		logger.Info("Window close intercepted: hiding window to tray as configured")
		wailsRuntime.WindowHide(ctx)
		return true
	}

	logger.Info("Window close proceeding: quitting application")
	return false
}

// onSecondInstanceLaunch is invoked when a second instance of the application attempts to start.
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	logger.Info(fmt.Sprintf("Second instance launch detected with args: %v", secondInstanceData.Args))
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	}
}

// GetRecentLogs returns recent log entries from the memory buffer.
func (a *App) GetRecentLogs() []logger.LogEntry {
	return logger.GetRecentLogs()
}

// ClearLogs clears the in-memory log buffer.
func (a *App) ClearLogs() {
	logger.ClearLogs()
}

// LogAction allows the frontend to log user UI interaction events directly into the Log Center.
func (a *App) LogAction(level string, message string, details string) {
	validLevels := map[string]bool{
		"":      true,
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLevels[level] {
		logger.Warn("Invalid log level from frontend", "level", level)
		level = "INFO"
	}

	const maxMessageLen = 1000
	const maxDetailsLen = 5000

	if len(message) > maxMessageLen {
		message = message[:maxMessageLen] + "... (truncated)"
	}
	if len(details) > maxDetailsLen {
		details = details[:maxDetailsLen] + "... (truncated)"
	}

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
	filtered := make([]disk.DiskInfo, 0, len(disks))
	for _, d := range disks {
		if hypervisor.IsDiskInVMSession(d.Device) {
			logger.Info("Filtering out disk currently engaged in active VM preview session", "disk", d.Device)
			continue
		}
		filtered = append(filtered, d)
	}
	logger.Info(fmt.Sprintf("Scanned removable storage drives, found %d device(s)", len(filtered)))
	return filtered, nil
}

// EjectDisk safely unmounts and ejects the target removable storage disk.
func (a *App) EjectDisk(targetDisk string) error {
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}
	if len(targetDisk) > 512 {
		return fmt.Errorf("target disk path too long (max 512 characters)")
	}

	logger.Info("Requesting explicit user-initiated safe ejection for selected disk", "disk", targetDisk)
	err := disk.SafeUserEjectDisk(targetDisk)
	if err != nil {
		logger.Error("Failed to eject target disk via explicit safety gate", "disk", targetDisk, "error", err)
		return err
	}
	disk.InvalidateDiskCache()
	logger.Info("Target disk safely ejected after explicit removable-disk validation", "disk", targetDisk)
	return nil
}

// BatchEjectDisks safely unmounts and ejects multiple removable storage disks.
func (a *App) BatchEjectDisks(targetDisks []string) BatchEjectResult {
	result := BatchEjectResult{
		Success: make([]string, 0, len(targetDisks)),
		Failed:  make(map[string]string),
	}

	if len(targetDisks) == 0 {
		return result
	}

	logger.Info("Requesting user-initiated safe batch ejection", "count", len(targetDisks), "disks", targetDisks)

	currentDisks, err := disk.GetRemovableDisks()
	if err != nil {
		logger.Error("Failed to fetch removable disk inventory for batch eject preflight", "error", err)
		for _, dev := range targetDisks {
			result.Failed[dev] = err.Error()
		}
		return result
	}

	diskMap := make(map[string]disk.DiskInfo, len(currentDisks))
	for _, d := range currentDisks {
		diskMap[d.Device] = d
	}

	validTargets := make([]string, 0, len(targetDisks))
	for _, dev := range targetDisks {
		devTrimmed := strings.TrimSpace(dev)
		if devTrimmed == "" || len(devTrimmed) > 512 {
			continue
		}
		candidate, exists := diskMap[devTrimmed]
		if !exists {
			result.Failed[devTrimmed] = "target disk is not present in the current removable-disk inventory"
			continue
		}
		if candidate.IsSystem {
			result.Failed[devTrimmed] = "CRITICAL: Safety block triggered! Disk is a system disk and cannot be ejected"
			continue
		}
		if !candidate.IsRemovable {
			result.Failed[devTrimmed] = "CRITICAL: Safety block triggered! Disk is not a removable disk"
			continue
		}
		validTargets = append(validTargets, devTrimmed)
	}

	for _, target := range validTargets {
		err := disk.EjectDisk(target)
		if err == nil {
			result.Success = append(result.Success, target)
			logger.Info("Target disk safely ejected", "disk", target)
		} else {
			result.Failed[target] = err.Error()
			logger.Error("Failed to eject target disk", "disk", target, "error", err)
		}
	}

	if len(result.Success) > 0 {
		time.Sleep(300 * time.Millisecond)
		disk.InvalidateDiskCache()
	}

	return result
}

// SelectIsoFiles opens a native multi-file open dialog for selecting Ventoy-supported system image files.
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

// SelectDirectory opens a native single folder selection dialog.
func (a *App) SelectDirectory(title string) (string, error) {
	if title == "" {
		title = "Select Folder"
	}
	defaultDir := ""
	titleLower := strings.ToLower(title)
	switch {
	case strings.Contains(titleLower, "ventoy"):
		defaultDir = env.GetVentoyDir()
	case strings.Contains(titleLower, "uniboot") || strings.Contains(titleLower, "firmware"):
		defaultDir = env.GetFirmwareDir()
	case strings.Contains(titleLower, "folder") || titleLower == "select folder":
		defaultDir = env.GetDataDir()
	}
	logger.Info("Opening native directory picker dialog", "title", title, "defaultDir", defaultDir)
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
	if err != nil {
		logger.Error("Failed to open directory picker", "error", err)
		return "", err
	}
	if dir != "" {
		logger.Info("Selected directory", "path", dir)
	}
	return dir, nil
}

// CalculateFileChecksum computes MD5, SHA256, or SHA512 hash for the specified image file.
func (a *App) CalculateFileChecksum(filePath string, algo string) (*utils.ChecksumResult, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}
	if len(filePath) > 4096 {
		return nil, fmt.Errorf("file path too long (max 4096 characters)")
	}

	validAlgos := map[string]bool{
		"md5":    true,
		"sha1":   true,
		"sha256": true,
		"sha384": true,
		"sha512": true,
	}
	algoLower := strings.ToLower(strings.TrimSpace(algo))
	if algoLower == "" {
		algoLower = "sha256"
	}
	if !validAlgos[algoLower] {
		return nil, fmt.Errorf("invalid checksum algorithm: %s (supported: md5, sha1, sha256, sha384, sha512)", algo)
	}

	logger.Info("Calculating file checksum", "filePath", filePath, "algorithm", algoLower)
	res, err := utils.CalculateFileChecksum(a.ctx, filePath, algoLower)
	if err != nil {
		logger.Error("Failed to calculate file checksum", "filePath", filePath, "algorithm", algoLower, "error", err)
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

	defaultFilename := fmt.Sprintf("unibootdesktop-log-%s.log", time.Now().Format("2006-01-02-150405"))
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
		return "", err
	}
	if filePath == "" {
		return "", nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		logger.Error("Failed to write exported logs to file", "path", filePath, "error", err)
		return "", err
	}

	logger.Info("Successfully exported log file", "path", filePath)
	return filePath, nil
}

// ValidateVentoyCli validates whether the target path contains a functional Ventoy installation executable.
func (a *App) ValidateVentoyCli(ventoyPath string) *installer.VentoyCliValidationResult {
	if strings.TrimSpace(ventoyPath) == "" {
		return &installer.VentoyCliValidationResult{
			Valid:   false,
			Code:    "path_empty",
			Message: "Ventoy CLI path cannot be empty",
		}
	}
	if len(ventoyPath) > 4096 {
		return &installer.VentoyCliValidationResult{
			Valid:   false,
			Code:    "path_too_long",
			Message: "Ventoy CLI path too long (max 4096 characters)",
		}
	}
	return installer.ValidateVentoyCli(ventoyPath)
}

// GetUniBootStatus returns the cross-platform ready status of native UniBoot engine.
func (a *App) GetUniBootStatus() *UniBootStatus {
	return &UniBootStatus{
		Ready:   true,
		Version: "1.0.0",
		Code:    "ready",
		Message: "UniBoot native cross-platform engine is ready",
	}
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
		a.emitEvent("iso-copy-progress", p)
	}
	batchProgressCb := func(p installer.BatchDeployProgress) {
		a.emitEvent("deploy-batch-progress", p)
	}
	return installer.DeployHybridModeBatchWithAllProgress(deployCtx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, expected, batchProgressCb)
}

// PreflightIsoCopy checks existing target filenames before any deployment writes begin.
func (a *App) PreflightIsoCopy(targetDisks []string, isoPaths []string) ([]installer.IsoCopyConflict, error) {
	type result struct {
		conflicts []installer.IsoCopyConflict
		err       error
	}

	ch := make(chan result, len(targetDisks))
	for _, targetDisk := range targetDisks {
		go func(td string) {
			if !disk.IsRealVentoyDisk(td) {
				ch <- result{}
				return
			}
			mountPoint, err := installer.ResolveMountPoint(td)
			if err != nil {
				ch <- result{err: fmt.Errorf("preflight failed to resolve target disk %s: %w", td, err)}
				return
			}
			current, err := installer.FindIsoCopyConflictsInDirectory(td, filepath.Join(mountPoint, "iso"), isoPaths)
			ch <- result{conflicts: current, err: err}
		}(targetDisk)
	}

	var conflicts []installer.IsoCopyConflict
	for range targetDisks {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		conflicts = append(conflicts, r.conflicts...)
	}
	return conflicts, nil
}

// DeployHybridModeBatchWithPlans executes Hybrid Mode using plans confirmed during preflight.
func (a *App) DeployHybridModeBatchWithPlans(targetDisks []string, fsType string, isoPaths []string, plans []installer.IsoCopyDiskPlan, expected []disk.DiskInfo) ([]*installer.DeployResult, error) {
	logger.Info("User confirmed batch Hybrid Mode deployment with ISO copy plans", "diskCount", len(targetDisks), "planCount", len(plans))
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	cfg, _ := config.Load()
	ventoyPath := ""
	if cfg != nil {
		ventoyPath = cfg.VentoyPath
	}
	progressCb := func(p installer.IsoCopyProgress) {
		a.emitEvent("iso-copy-progress", p)
	}
	batchProgressCb := func(p installer.BatchDeployProgress) {
		a.emitEvent("deploy-batch-progress", p)
	}
	return installer.DeployHybridModeBatchWithIsoPlans(deployCtx, targetDisks, fsType, ventoyPath, isoPaths, plans, progressCb, expected, batchProgressCb)
}

// DeployCloudMode triggers Cloud Mode (Cloud Pure Mode) with customizable file system.
func (a *App) DeployCloudMode(targetDisk string, fsType string, expected disk.DiskInfo) (*installer.DeployResult, error) {
	logger.Info("User confirmed Cloud Mode boot disk creation", "disk", targetDisk, "fs", fsType)
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	progressCallback := func(progress int) {
		a.emitEvent("cloud-deploy-progress", progress)
	}

	res, err := installer.DeployCloudModeWithExpectedDisk(deployCtx, targetDisk, fsType, expected, progressCallback)
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

	progressCallback := func(progress installer.BatchDeployProgress) {
		a.emitEvent("deploy-batch-progress", progress)
		a.emitEvent("cloud-deploy-batch-progress", progress)
	}

	results, err := installer.DeployCloudModeBatchWithExpectedDisks(deployCtx, targetDisks, fsType, expected, progressCallback)
	return results, err
}

// CheckQEMU returns QEMU detection metadata for backward compatibility.
func (a *App) CheckQEMU() *hypervisor.VMStatus {
	best := hypervisor.GetManager().DetectBest()
	if best != nil && best.Installed {
		return best
	}
	for _, drv := range hypervisor.GetManager().DetectAll() {
		if drv.Type == hypervisor.TypeQEMU {
			return drv
		}
	}
	return &hypervisor.VMStatus{
		Type:       hypervisor.TypeQEMU,
		Name:       "QEMU",
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   1,
		CanBootRaw: false,
	}
}

// DetectHypervisors returns status of all installed virtual machine engines.
func (a *App) DetectHypervisors() []*hypervisor.VMStatus {
	return hypervisor.GetManager().DetectAll()
}

// DetectBestHypervisor returns the highest priority available virtual machine status.
func (a *App) DetectBestHypervisor() *hypervisor.VMStatus {
	return hypervisor.GetManager().DetectBest()
}

// LaunchQEMU triggers virtual machine test instance using highest priority available hypervisor.
func (a *App) LaunchQEMU(targetDisk string) error {
	if err := disk.ValidateUserEjectTarget(targetDisk); err != nil {
		return fmt.Errorf("unsafe VM target disk: %w", err)
	}
	logger.Info("Requesting hypervisor preview test launch", "disk", targetDisk)
	err := hypervisor.GetManager().LaunchBest(a.ctx, targetDisk, hypervisor.BootModeAuto)
	if err != nil {
		logger.Error("Failed to launch hypervisor preview test", "disk", targetDisk, "error", err)
		return err
	}
	logger.Info("Hypervisor preview test launched successfully", "disk", targetDisk)
	disk.InvalidateDiskCache()
	if a.ctx != nil {
		a.emitEvent("disk-list-changed")
	}
	return nil
}

// LaunchVM launches a specified or best available virtual machine with boot mode (uefi, bios, auto).
func (a *App) LaunchVM(targetDisk string, vmType string, bootMode string) error {
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}
	if len(targetDisk) > 512 {
		return fmt.Errorf("target disk path too long")
	}

	validVMTypes := map[string]bool{
		"":           true,
		"auto":       true,
		"qemu":       true,
		"utm":        true,
		"vmware":     true,
		"virtualbox": true,
	}
	vmTypeLower := strings.ToLower(strings.TrimSpace(vmType))
	if !validVMTypes[vmTypeLower] {
		return fmt.Errorf("invalid VM type: %s (supported: auto, qemu, utm, vmware, virtualbox)", vmType)
	}

	validBootModes := map[string]bool{
		"":     true,
		"auto": true,
		"uefi": true,
		"bios": true,
	}
	bootModeLower := strings.ToLower(strings.TrimSpace(bootMode))
	if !validBootModes[bootModeLower] {
		return fmt.Errorf("invalid boot mode: %s (supported: auto, uefi, bios)", bootMode)
	}

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
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}
	if len(targetDisk) > 512 {
		return fmt.Errorf("target disk path too long")
	}
	if err := disk.ValidateUserEjectTarget(targetDisk); err != nil {
		return fmt.Errorf("unsafe VM target disk: %w", err)
	}

	if cfg.CpuCores < 0 || cfg.CpuCores > 256 {
		return fmt.Errorf("invalid CPU cores: %d (must be 0-256)", cfg.CpuCores)
	}
	if cfg.MemoryMB < 0 || cfg.MemoryMB > 1048576 {
		return fmt.Errorf("invalid memory: %d MB (must be 0-1048576)", cfg.MemoryMB)
	}

	if cfg.BootMode == "" {
		cfg.BootMode = hypervisor.BootModeAuto
	}
	if cfg.CpuCores <= 0 {
		cfg.CpuCores = hypervisor.GetRecommendedVCPUs()
	}
	if cfg.MemoryMB <= 0 {
		cfg.MemoryMB = hypervisor.GetRecommendedVMMemoryMB()
	}

	var err error
	if vmType == "" || vmType == "auto" {
		err = hypervisor.GetManager().LaunchBestConfigured(a.ctx, targetDisk, cfg)
	} else {
		logger.Info("Requesting specified hypervisor preview test launch with VMConfig", "disk", targetDisk, "vmType", vmType, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)
		err = hypervisor.GetManager().LaunchSpecifiedConfigured(a.ctx, targetDisk, hypervisor.HypervisorType(vmType), cfg)
	}
	if err != nil {
		logger.Error("Failed to launch hypervisor with VMConfig", "disk", targetDisk, "vmType", vmType, "error", err)
		return err
	}
	logger.Info("Specified hypervisor test launched successfully with VMConfig", "disk", targetDisk, "vmType", vmType)
	disk.InvalidateDiskCache()
	if a.ctx != nil {
		a.emitEvent("disk-list-changed")
	}
	return nil
}

// StopVM manually stops any active virtual machine simulation session.
func (a *App) StopVM() error {
	logger.Info("User manually requested VM simulation session stop")
	hypervisor.StopActiveVMSession()
	hypervisor.GetManager().CleanupAllUnmountedDisks()
	if a.ctx != nil {
		a.emitEvent("vm-session-ended")
	}
	return nil
}

// CheckUpdate returns GitHub release update metadata.
func (a *App) CheckUpdate() *updater.UpdateStatus {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return updater.CheckUpdate(ctx)
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
	if cfg.VentoyPath == "" {
		cfg.VentoyPath = env.GetVentoyDir()
	}
	if cfg.UniBootPath == "" {
		cfg.UniBootPath = env.GetFirmwareDir()
	}

	if cfg.GithubProxy != "" && cfg.GithubProxy != "direct" {
		proxyURL, err := url.Parse(cfg.GithubProxy)
		if err != nil {
			logger.Error("Invalid GitHub proxy URL format", "proxy", cfg.GithubProxy, "error", err)
			return fmt.Errorf("invalid GitHub proxy URL: %w", err)
		}
		if proxyURL.Scheme != "http" && proxyURL.Scheme != "https" {
			logger.Error("GitHub proxy must use http or https scheme", "scheme", proxyURL.Scheme)
			return fmt.Errorf("GitHub proxy must use http:// or https:// (got %s://)", proxyURL.Scheme)
		}
		if proxyURL.Host == "" {
			logger.Error("GitHub proxy URL missing host")
			return fmt.Errorf("GitHub proxy URL must have a valid host")
		}
	}

	if cfg.ProxyPort < 0 || cfg.ProxyPort > 65535 {
		logger.Error("Invalid proxy port", "port", cfg.ProxyPort)
		return fmt.Errorf("proxy port must be between 0 and 65535 (got %d)", cfg.ProxyPort)
	}

	if cfg.ProxyProtocol == "direct" || cfg.ProxyHost == "" {
		cfg.ProxyPort = 0
	} else if cfg.ProxyPort > 0 && cfg.ProxyHost == "" {
		logger.Error("Proxy host is empty but port is set", "port", cfg.ProxyPort)
		return fmt.Errorf("proxy host cannot be empty when proxy port is configured")
	}

	if cfg.Language == "" {
		cfg.Language = "auto"
	}

	if cfg.Theme != "light" && cfg.Theme != "dark" && cfg.Theme != "system" {
		cfg.Theme = "system"
	}

	validFileSystems := map[string]bool{
		"":      true,
		"exFAT": true,
		"NTFS":  true,
		"FAT32": true,
		"ext4":  true,
	}
	if !validFileSystems[cfg.FileSystem] {
		logger.Error("Invalid file system type", "fileSystem", cfg.FileSystem)
		return fmt.Errorf("invalid file system: %s (must be one of: exFAT, NTFS, FAT32, ext4)", cfg.FileSystem)
	}

	if cfg.Mode != "" && cfg.Mode != "cloud" && cfg.Mode != "hybrid" {
		logger.Error("Invalid mode", "mode", cfg.Mode)
		return fmt.Errorf("invalid mode: %s (must be 'cloud' or 'hybrid')", cfg.Mode)
	}

	validProxyProtocols := map[string]bool{
		"":       true,
		"direct": true,
		"http":   true,
		"https":  true,
		"socks4": true,
		"socks5": true,
	}
	if !validProxyProtocols[cfg.ProxyProtocol] {
		logger.Error("Invalid proxy protocol", "protocol", cfg.ProxyProtocol)
		return fmt.Errorf("invalid proxy protocol: %s (must be one of: direct, http, https, socks4, socks5)", cfg.ProxyProtocol)
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
	firmware.SetCustomUniBootDir(cfg.UniBootPath)
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

// GetSystemInfo returns system environment and runtime diagnostic metadata.
func (a *App) GetSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		AppName:   "UniBootDesktop",
		Version:   env.GitTag,
		Commit:    env.CommitHash,
		BuildTime: env.BuildTime,
		DataDir:   env.GetDataDir(),
		ConfigDir: env.GetConfigDir(),
	}
}

// GetHelloInfo returns structured hello greeting and runtime environment details.
func (a *App) GetHelloInfo() *HelloInfo {
	return &HelloInfo{
		Greeting:  "Hello World From UniBootDesktop!",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// Greet returns a friendly greeting for demonstration.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello %s, Welcome to UniBootDesktop!", name)
}

// TestNetwork checks network connectivity and latency against target endpoint.
func (a *App) TestNetwork(targetURL string) *NetworkTestResult {
	if targetURL == "" {
		targetURL = "https://api.github.com"
	}
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &NetworkTestResult{
			Connected: false,
			LatencyMs: latency,
			TargetURL: targetURL,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	return &NetworkTestResult{
		Connected: resp.StatusCode >= 200 && resp.StatusCode < 400,
		LatencyMs: latency,
		TargetURL: targetURL,
	}
}

// PerformGuiUpdate performs background download and staging of the latest GUI release.
func (a *App) PerformGuiUpdate() (*updater.GuiUpdateResult, error) {
	if !a.mu.TryLock() {
		return nil, fmt.Errorf("update is already in progress")
	}
	defer a.mu.Unlock()

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	proxyPrefix := ""
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		proxyPrefix = cfg.GithubProxy
	}
	if proxyPrefix == "" {
		proxyPrefix = env.GithubProxy()
	}

	progressCallback := func(p updater.UpdateProgress) {
		if a.ctx != nil {
			a.emitEvent("gui-update-progress", p)
		}
	}

	return updater.PerformGuiUpdate(ctx, proxyPrefix, progressCallback)
}

// OpenBrowserURL opens the target URL in the user's default system browser.
func (a *App) OpenBrowserURL(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		logger.Error("Failed to parse URL for browser open", "url", targetURL, "error", err)
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		logger.Warn("Blocked non-HTTPS URL from being opened in browser", "url", targetURL, "scheme", u.Scheme)
		return fmt.Errorf("only HTTPS URLs are allowed for security reasons (got: %s://)", u.Scheme)
	}

	if u.Host == "" {
		logger.Error("URL has no host", "url", targetURL)
		return fmt.Errorf("invalid URL: missing host")
	}

	if isPrivateOrLocalIP(u.Hostname()) {
		logger.Warn("Blocked private/local IP address from being opened", "url", targetURL, "host", u.Hostname())
		return fmt.Errorf("access to private/local IP addresses is not allowed for security reasons")
	}

	logger.Info("Opening URL in system browser", "url", targetURL)
	if a.ctx != nil && a.ctx.Value("frontend") != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, targetURL)
	}
	return nil
}

// OpenURL opens the specified URL in the native desktop browser (backward compatible).
func (a *App) OpenURL(rawURL string) {
	if rawURL != "" {
		_ = a.OpenBrowserURL(rawURL)
	}
}

// OpenAboutModal emits an event to the frontend to trigger the About dialog.
func (a *App) OpenAboutModal() {
	if a.ctx != nil {
		a.emitEvent("open-about-modal")
	}
}

// ReloadAppMenu rebuilds and updates the native application menu with the specified language.
func (a *App) ReloadAppMenu(lang string) error {
	if a.ctx == nil || a.ctx.Value("frontend") == nil {
		return nil
	}
	appMenu := BuildAppMenu(a, lang)
	wailsRuntime.MenuSetApplicationMenu(a.ctx, appMenu)
	wailsRuntime.MenuUpdateApplicationMenu(a.ctx)
	return nil
}

func isPrivateOrLocalIP(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return true
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	}
	if host == "localhost" || host == "localhost." ||
		strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

// IsPrivileged returns true if the app process or worker currently possesses administrator or root privileges.
func (a *App) IsPrivileged() bool {
	return privilege.IsElevated()
}

// RequestPrivilegeElevation prompts the user for administrator privileges across operating systems.
func (a *App) RequestPrivilegeElevation() (bool, error) {
	if privilege.IsElevated() {
		disk.InvalidateDiskCache()
		return true, nil
	}

	prompt := "UniBootDesktop requires administrator privileges to access raw storage devices and verify boot partitions."
	_, err := privilege.StartOrConnectWorker(prompt)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "canceled") || strings.Contains(errStr, "cancelled") || strings.Contains(errStr, "user declined") || strings.Contains(errStr, "-128") {
			logger.Info("User dismissed privilege elevation prompt")
			return false, err
		}

		logger.Warn("Failed to start privileged worker", "error", err)
		return false, err
	}

	privilege.ResetElevationCache()
	disk.InvalidateDiskCache()
	logger.Info("Administrator privilege successfully granted by user, disk cache invalidated")
	return true, nil
}

var execCommand = exec.Command

// RestartApp gracefully quits and restarts the application or applies pending updates.
func (a *App) RestartApp() error {
	pending, err := updater.GetPendingUpdate(env.GetDataDir())
	if err == nil && pending != nil && pending.ScriptPath != "" {
		if _, err := os.Stat(pending.ScriptPath); err == nil {
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = execCommand(pending.Shell, "/c", pending.ScriptPath, strconv.Itoa(os.Getpid()), pending.Target, pending.Staged, filepath.Dir(pending.ScriptPath))
			} else {
				cmd = execCommand(pending.Shell, pending.ScriptPath, strconv.Itoa(os.Getpid()), pending.Target, pending.Staged, filepath.Dir(pending.ScriptPath))
			}
			detachProcess(cmd)
			if err := cmd.Start(); err == nil {
				if a.ctx != nil && a.ctx.Value("frontend") != nil {
					wailsRuntime.Quit(a.ctx)
				}
				return nil
			}
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := execCommand(exe, os.Args[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart application: %w", err)
	}

	if a.ctx != nil && a.ctx.Value("frontend") != nil {
		wailsRuntime.Quit(a.ctx)
	}
	return nil
}
