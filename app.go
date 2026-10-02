// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// execCommand is a package-level variable to allow tests to mock OS process creation.
var execCommand = func(name string, args ...string) *exec.Cmd {
	return newDetachedCmd(name, args...)
}

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.Mutex
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
	a.ctx, a.cancel = context.WithCancel(ctx)
	logger.SetWailsContext(a.ctx)
	logger.Info(fmt.Sprintf("UniBootDesktop Wails GUI runtime started successfully (%s/%s)", runtime.GOOS, runtime.GOARCH))
	if err := env.EnsureAppDirs(); err != nil {
		logger.Warn("Failed to ensure default app directories exist", "error", err)
	}
	if cfg, err := config.Load(); err == nil && cfg != nil && cfg.UniBootPath != "" {
		firmware.SetCustomUniBootDir(cfg.UniBootPath)
	}
	disk.StartHotplugMonitor(a.ctx, func() {
		logger.Info("Removable disk change detected, refreshing drive list")
		wailsRuntime.EventsEmit(a.ctx, "disk-list-changed")
	})

	hypervisor.RegisterVMExitHandler(func(targetDisk string, vmErr error) {
		logger.Info("Hypervisor VM session terminated, bouncing UI back and refreshing disk state", "disk", targetDisk)
		disk.InvalidateDiskCache()
		if a.ctx != nil {
			wailsRuntime.WindowUnminimise(a.ctx)
			wailsRuntime.WindowShow(a.ctx)
			wailsRuntime.EventsEmit(a.ctx, "disk-list-changed")
			wailsRuntime.EventsEmit(a.ctx, "vm-session-ended", map[string]interface{}{
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
	logger.Info("UniBootDesktop Wails GUI runtime shutting down; safe policy is to only remount tracked VM disks, never eject arbitrary USB media")

	if a.cancel != nil {
		a.cancel()
	}

	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		hypervisor.GetManager().CleanupAllUnmountedDisks()
	}()

	select {
	case <-cleanupDone:
		logger.Info("Tracked hypervisor disk cleanup completed before shutdown timeout")
	case <-time.After(5 * time.Second):
		logger.Warn("Tracked hypervisor disk cleanup hit shutdown timeout; continuing app exit without forcing arbitrary USB ejection")
	}

	// Cleanly disconnect and terminate active privileged worker session
	if client := privilege.GetActiveWorkerClient(); client != nil {
		_ = client.Close()
		privilege.SetActiveWorkerClient(nil)
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
	// 验证日志级别
	validLevels := map[string]bool{
		"":      true, // 空字符串默认为INFO
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLevels[level] {
		logger.Warn("Invalid log level from frontend", "level", level)
		level = "INFO"
	}

	// 限制消息和详情长度，防止日志洪水攻击
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
	// 验证输入不为空
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}

	// 验证路径长度，防止过长路径攻击
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

// BatchEjectResult describes the outcome of a concurrent multi-disk ejection.
type BatchEjectResult struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"`
}

// BatchEjectDisks safely unmounts and ejects multiple removable storage disks.
// Disks are verified up-front and ejected sequentially to prevent OS disk arbitration daemon
// (e.g. macOS diskarbitrationd) lock contention and timeouts.
func (a *App) BatchEjectDisks(targetDisks []string) BatchEjectResult {
	result := BatchEjectResult{
		Success: make([]string, 0, len(targetDisks)),
		Failed:  make(map[string]string),
	}

	if len(targetDisks) == 0 {
		return result
	}

	logger.Info("Requesting user-initiated safe batch ejection", "count", len(targetDisks), "disks", targetDisks)

	// Step 1: Pre-validate all target disks up-front against current removable inventory
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

	// Step 2: Eject valid targets sequentially to avoid OS disk arbitration collisions
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
		// Allow macOS/Windows kernel brief window to complete IOKit/device node teardown
		time.Sleep(300 * time.Millisecond)
		disk.InvalidateDiskCache()
	}

	return result
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

	var filters []wailsRuntime.FileFilter
	if runtime.GOOS == "darwin" {
		// On macOS (Cocoa), Wails translates patterns by stripping "*." and passing extensions to UTType / setAllowedFileTypes.
		// Specifying "*.*" creates an invalid "*" extension that causes NSOpenPanel to throw an Objective-C exception or malfunction.
		// On macOS we provide clean valid extensions, omitting "*.*".
		filters = []wailsRuntime.FileFilter{
			{
				DisplayName: ventoyFilter,
				Pattern:     "*.iso;*.img;*.wim;*.vhd;*.vhdx;*.vti;*.efi;*.bin;*.xz;*.gz;*.raw",
			},
		}
	} else {
		filters = []wailsRuntime.FileFilter{
			{
				DisplayName: ventoyFilter,
				Pattern:     "*.iso;*.wim;*.img;*.vhd;*.vhdx;*.vti;*.efi;*.bin;*.xz;*.gz;*.raw",
			},
			{
				DisplayName: allFilter,
				Pattern:     "*.*",
			},
		}
	}

	paths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
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
	// 验证输入
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// 验证路径长度
	if len(filePath) > 4096 {
		return nil, fmt.Errorf("file path too long (max 4096 characters)")
	}

	// 验证算法参数
	validAlgos := map[string]bool{
		"md5":    true,
		"sha1":   true,
		"sha256": true,
		"sha384": true,
		"sha512": true,
	}
	algoLower := strings.ToLower(strings.TrimSpace(algo))
	if algoLower == "" {
		algoLower = "sha256" // 默认使用SHA256
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
	var saveFilters []wailsRuntime.FileFilter
	if runtime.GOOS == "darwin" {
		saveFilters = []wailsRuntime.FileFilter{
			{
				DisplayName: logFilter,
				Pattern:     "*.log",
			},
			{
				DisplayName: textFilter,
				Pattern:     "*.txt",
			},
		}
	} else {
		saveFilters = []wailsRuntime.FileFilter{
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
		}
	}

	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters:         saveFilters,
	})
	if err != nil {
		logger.Error("Failed to open save file dialog for log export", "error", err)
		return "", fmt.Errorf("open save file dialog: %w", err)
	}
	if filePath == "" {
		return "", nil // User cancelled
	}
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
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
	// 验证路径输入
	if strings.TrimSpace(ventoyPath) == "" {
		return &installer.VentoyCliValidationResult{
			Valid:   false,
			Code:    "path_empty",
			Message: "Ventoy CLI path cannot be empty",
		}
	}

	// 验证路径长度
	if len(ventoyPath) > 4096 {
		return &installer.VentoyCliValidationResult{
			Valid:   false,
			Code:    "path_too_long",
			Message: "Ventoy CLI path too long (max 4096 characters)",
		}
	}

	return installer.ValidateVentoyCli(ventoyPath)
}

// UniBootStatus represents the lightweight cross-platform ready status of native UniBoot engine.
type UniBootStatus struct {
	Ready   bool   `json:"ready"`
	Version string `json:"version"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
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
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	batchProgressCb := func(p installer.BatchDeployProgress) {
		wailsRuntime.EventsEmit(a.ctx, "deploy-batch-progress", p)
	}
	return installer.DeployHybridModeBatchWithAllProgress(deployCtx, targetDisks, fsType, ventoyPath, isoPaths, progressCb, expected, batchProgressCb)
}

// PreflightIsoCopy checks existing target filenames before any deployment writes begin.
// Each disk is checked concurrently to minimize total latency in batch scenarios.
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
		wailsRuntime.EventsEmit(a.ctx, "iso-copy-progress", p)
	}
	batchProgressCb := func(p installer.BatchDeployProgress) {
		wailsRuntime.EventsEmit(a.ctx, "deploy-batch-progress", p)
	}
	return installer.DeployHybridModeBatchWithIsoPlans(deployCtx, targetDisks, fsType, ventoyPath, isoPaths, plans, progressCb, expected, batchProgressCb)
}

// DeployCloudMode triggers Cloud Mode (Cloud Pure Mode) with customizable file system.
func (a *App) DeployCloudMode(targetDisk string, fsType string, expected disk.DiskInfo) (*installer.DeployResult, error) {
	logger.Info("User confirmed Cloud Mode boot disk creation", "disk", targetDisk, "fs", fsType)
	deployCtx := a.initDeployContext()
	defer a.clearDeployContext()

	// Create progress callback to emit real-time progress events
	progressCallback := func(progress int) {
		wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-progress", progress)
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

	// Create batch progress callback to emit real-time progress events with disk-level details
	progressCallback := func(progress installer.BatchDeployProgress) {
		wailsRuntime.EventsEmit(a.ctx, "deploy-batch-progress", progress)
		wailsRuntime.EventsEmit(a.ctx, "cloud-deploy-batch-progress", progress)
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

// LaunchQEMU triggers virtual machine test instance using highest priority available hypervisor (QEMU, UTM, VMware, VirtualBox).
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
		wailsRuntime.EventsEmit(a.ctx, "disk-list-changed")
	}
	return nil
}

// LaunchVM launches a specified or best available virtual machine with boot mode (uefi, bios, auto).
func (a *App) LaunchVM(targetDisk string, vmType string, bootMode string) error {
	// 验证targetDisk
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}
	if len(targetDisk) > 512 {
		return fmt.Errorf("target disk path too long")
	}

	// 验证vmType
	validVMTypes := map[string]bool{
		"":           true, // 空表示auto
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

	// 验证bootMode
	validBootModes := map[string]bool{
		"":     true, // 空表示auto
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
	// 验证targetDisk
	if strings.TrimSpace(targetDisk) == "" {
		return fmt.Errorf("target disk path cannot be empty")
	}
	if len(targetDisk) > 512 {
		return fmt.Errorf("target disk path too long")
	}
	if err := disk.ValidateUserEjectTarget(targetDisk); err != nil {
		return fmt.Errorf("unsafe VM target disk: %w", err)
	}

	// 验证VMConfig参数范围
	if cfg.CpuCores < 0 || cfg.CpuCores > 256 {
		return fmt.Errorf("invalid CPU cores: %d (must be 0-256)", cfg.CpuCores)
	}
	if cfg.MemoryMB < 0 || cfg.MemoryMB > 1048576 { // 最大1TB
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
		wailsRuntime.EventsEmit(a.ctx, "disk-list-changed")
	}
	return nil
}

// StopVM manually stops any active virtual machine simulation session, remounts target disks, and resets UI state.
func (a *App) StopVM() error {
	logger.Info("User manually requested VM simulation session stop")
	hypervisor.StopActiveVMSession()
	hypervisor.GetManager().CleanupAllUnmountedDisks()
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "vm-session-ended")
	}
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
	if cfg.VentoyPath == "" {
		cfg.VentoyPath = env.GetVentoyDir()
	}
	if cfg.UniBootPath == "" {
		cfg.UniBootPath = env.GetFirmwareDir()
	}

	// Validate GithubProxy URL format
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

	// Validate ProxyPort range (1-65535)
	if cfg.ProxyPort < 0 || cfg.ProxyPort > 65535 {
		logger.Error("Invalid proxy port", "port", cfg.ProxyPort)
		return fmt.Errorf("proxy port must be between 0 and 65535 (got %d)", cfg.ProxyPort)
	}

	// Normalize proxy configuration: if direct or host is empty, reset port to 0
	if cfg.ProxyProtocol == "direct" || cfg.ProxyHost == "" {
		cfg.ProxyPort = 0
	} else if cfg.ProxyPort > 0 && cfg.ProxyHost == "" {
		logger.Error("Proxy host is empty but port is set", "port", cfg.ProxyPort)
		return fmt.Errorf("proxy host cannot be empty when proxy port is configured")
	}

	// Validate Language enum (supporting all 52 internationalized locales + auto)
	validLanguages := map[string]bool{
		"auto":    true,
		"zh-CN":   true,
		"en-US":   true,
		"zh-TW":   true,
		"ja-JP":   true,
		"ko-KR":   true,
		"de-DE":   true,
		"fr-FR":   true,
		"es-ES":   true,
		"es-LA":   true,
		"ru-RU":   true,
		"pt-BR":   true,
		"pt-PT":   true,
		"it-IT":   true,
		"tr-TR":   true,
		"pl-PL":   true,
		"vi-VN":   true,
		"ar-SA":   true,
		"ur-PK":   true,
		"az-AZ":   true,
		"da-DK":   true,
		"ka-GE":   true,
		"fa-IR":   true,
		"sl-SI":   true,
		"oc-FR":   true,
		"cs-CZ":   true,
		"sk-SK":   true,
		"bn-BD":   true,
		"hi-IN":   true,
		"nl-NL":   true,
		"ro-RO":   true,
		"hr-HR":   true,
		"hu-HU":   true,
		"sr-Latn": true,
		"sr-Cyrl": true,
		"th-TH":   true,
		"lt-LT":   true,
		"mk-MK":   true,
		"he-IL":   true,
		"id-ID":   true,
		"nb-NO":   true,
		"uk-UA":   true,
		"el-GR":   true,
		"sv-SE":   true,
		"bg-BG":   true,
		"hy-AM":   true,
		"fi-FI":   true,
		"gl-ES":   true,
		"ca-ES":   true,
		"ta-IN":   true,
		"be-BY":   true,
		"ml-IN":   true,
		"et-EE":   true,
	}
	if !validLanguages[cfg.Language] {
		logger.Error("Invalid language code", "language", cfg.Language)
		return fmt.Errorf("invalid language code: %s", cfg.Language)
	}

	// Normalize theme preference (support "dark", "light", "system", defaulting to "system")
	if cfg.Theme != "light" && cfg.Theme != "dark" && cfg.Theme != "system" {
		cfg.Theme = "system"
	}

	// Validate FileSystem enum
	validFileSystems := map[string]bool{
		"exFAT": true,
		"NTFS":  true,
		"FAT32": true,
		"ext4":  true,
	}
	if !validFileSystems[cfg.FileSystem] {
		logger.Error("Invalid file system type", "fileSystem", cfg.FileSystem)
		return fmt.Errorf("invalid file system: %s (must be one of: exFAT, NTFS, FAT32, ext4)", cfg.FileSystem)
	}

	// Validate Mode enum
	if cfg.Mode != "cloud" && cfg.Mode != "hybrid" {
		logger.Error("Invalid mode", "mode", cfg.Mode)
		return fmt.Errorf("invalid mode: %s (must be 'cloud' or 'hybrid')", cfg.Mode)
	}

	// Validate ProxyProtocol enum
	validProxyProtocols := map[string]bool{
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

// GetVentoyReleaseInfo queries the latest official Ventoy GitHub release metadata.
func (a *App) GetVentoyReleaseInfo() (*installer.VentoyReleaseInfo, error) {
	cfg, _ := config.Load()
	proxy := ""
	ventoyPath := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
		ventoyPath = cfg.VentoyPath
	}
	if ventoyPath == "" {
		ventoyPath = env.GetVentoyDir()
	}
	return installer.FetchLatestVentoyRelease(a.ctx, proxy, ventoyPath)
}

// DownloadVentoyRelease downloads and extracts the official Ventoy toolchain for current platform into the designated Ventoy directory.
func (a *App) DownloadVentoyRelease() (*installer.VentoyReleaseInfo, error) {
	cfg, _ := config.Load()
	proxy := ""
	ventoyPath := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
		ventoyPath = cfg.VentoyPath
	}
	if ventoyPath == "" {
		ventoyPath = env.GetVentoyDir()
	}
	info, err := installer.DownloadAndExtractVentoy(a.ctx, proxy, ventoyPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil && cfg.VentoyPath == "" {
		cfg.VentoyPath = ventoyPath
		_ = cfg.Save()
	}
	return info, nil
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
	if !a.mu.TryLock() {
		return nil, fmt.Errorf("update already in progress")
	}
	defer a.mu.Unlock()

	cfg, _ := config.Load()
	proxy := ""
	if cfg != nil {
		proxy = cfg.GithubProxy
	}
	progressCb := func(p updater.UpdateProgress) {
		wailsRuntime.EventsEmit(a.ctx, "gui-update-progress", p)
	}
	return updater.PerformGuiUpdate(a.ctx, proxy, progressCb)
}

// OpenBrowserURL opens the target URL in the user's default system browser.
// Only HTTPS URLs are allowed for security (prevents file://, javascript:, etc.)
func (a *App) OpenBrowserURL(targetURL string) error {
	// Parse and validate URL
	u, err := url.Parse(targetURL)
	if err != nil {
		logger.Error("Failed to parse URL for browser open", "url", targetURL, "error", err)
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Protocol whitelist: only allow HTTPS for security
	// Prevents: file:/// (local file access), javascript: (XSS), data: (data URI), etc.
	if u.Scheme != "https" {
		logger.Warn("Blocked non-HTTPS URL from being opened in browser", "url", targetURL, "scheme", u.Scheme)
		return fmt.Errorf("only HTTPS URLs are allowed for security reasons (got: %s://)", u.Scheme)
	}

	// Validate host is not empty
	if u.Host == "" {
		logger.Error("URL has no host", "url", targetURL)
		return fmt.Errorf("invalid URL: missing host")
	}

	// SSRF防护: 阻止内网IP地址和本地主机访问
	if isPrivateOrLocalIP(u.Hostname()) {
		logger.Warn("Blocked private/local IP address from being opened", "url", targetURL, "host", u.Hostname())
		return fmt.Errorf("access to private/local IP addresses is not allowed for security reasons")
	}

	logger.Info("Opening URL in system browser", "url", targetURL)
	wailsRuntime.BrowserOpenURL(a.ctx, targetURL)
	return nil
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

// isPrivateOrLocalIP reports whether host is loopback, link-local, or RFC1918/ULA.
// host should come from url.Hostname() (no port). IPv4-mapped addresses are
// checked via net.IP.IsPrivate, not via ip[0] on the 16-byte form.
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
		// If user actively cancelled the prompt, do not pop up a second prompt
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

// emitEvent safely emits a Wails event if the app context is ready.
func (a *App) emitEvent(eventName string, optionalData ...interface{}) {
	if a.ctx == nil {
		return
	}
	if len(optionalData) > 0 {
		wailsRuntime.EventsEmit(a.ctx, eventName, optionalData...)
	} else {
		wailsRuntime.EventsEmit(a.ctx, eventName)
	}
}

// beforeClose is called when the user attempts to close the application window.
// It returns true to prevent the default close behavior (e.g. to minimize to tray instead).
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return false
	}
	if cfg.EnableTray && cfg.CloseAction == "minimize_to_tray" {
		wailsRuntime.WindowHide(ctx)
		return true
	}
	return false
}

// onSecondInstanceLaunch is called when a second instance of the application is launched.
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	if a.ctx != nil {
		wailsRuntime.WindowUnminimise(a.ctx)
		wailsRuntime.WindowShow(a.ctx)
	}
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
	GoVersion string `json:"goVersion"`
}

// GetSystemInfo returns comprehensive runtime and environment info.
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
		ConfigDir: filepath.Dir(env.GetGlobalConfigPath()),
	}
}

// GetHelloInfo returns a simple greeting and runtime context for demos.
func (a *App) GetHelloInfo() *HelloInfo {
	return &HelloInfo{
		Greeting:  "Hello World, Welcome to UniBootDesktop!",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
}

// Greet returns a personalised greeting string (used in tests and demos).
func (a *App) Greet(name string) string {
	if name == "" {
		return "Hello World, Welcome to UniBootDesktop!"
	}
	return fmt.Sprintf("Hello %s, Welcome to UniBootDesktop!", name)
}

// TestNetwork performs a quick connectivity probe to the specified URL and returns latency.
func (a *App) TestNetwork(targetURL string) *NetworkTestResult {
	if targetURL == "" {
		targetURL = "https://www.google.com"
	}
	start := time.Now()
	result := &NetworkTestResult{TargetURL: targetURL}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		result.Error = fmt.Sprintf("invalid URL: %v", err)
		return result
	}
	host := parsed.Hostname()
	if host == "" {
		result.Error = "empty host"
		return result
	}
	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			port = "80"
		}
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 5*time.Second)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	conn.Close()
	result.Connected = true
	result.LatencyMs = time.Since(start).Milliseconds()
	return result
}

// OpenURL opens the given URL in the user's default system browser.
func (a *App) OpenURL(rawURL string) {
	if a.ctx == nil {
		return
	}
	wailsRuntime.BrowserOpenURL(a.ctx, rawURL)
}

// RestartApp relaunches the application binary as a detached child and then quits the current instance.
func (a *App) RestartApp() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		execPath = exe
	}
	cmd := execCommand(execPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new instance: %w", err)
	}
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
	return nil
}
