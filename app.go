// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/internal/sysinfo"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
	"github.com/snowdreamtech/unigodesktop/pkg/disk"
	"github.com/snowdreamtech/unigodesktop/pkg/hypervisor"
	"github.com/snowdreamtech/unigodesktop/pkg/privilege"
	"github.com/snowdreamtech/unigodesktop/pkg/updater"
	"github.com/snowdreamtech/unigodesktop/pkg/utils"
)

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

// App struct manages Wails GUI lifecycle and frontend bound APIs.
type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// emitEvent safely emits a Wails event, guarding against unit test contexts and uninitialized runtime.
func (a *App) emitEvent(eventName string, optionalData ...interface{}) {
	if a.ctx == nil || a.ctx.Value("frontend") == nil {
		return
	}
	if len(optionalData) > 0 {
		wailsRuntime.EventsEmit(a.ctx, eventName, optionalData...)
	} else {
		wailsRuntime.EventsEmit(a.ctx, eventName)
	}
}

// startup is called when the Wails application starts up.
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	logger.SetWailsContext(ctx)

	if cfg, err := config.Load(); err == nil && cfg != nil {
		if strings.HasPrefix(strings.ToLower(cfg.Language), "zh") {
			privilege.SetDefaultElevationPrompt("UniGoDesktop éœ€è¦ç®¡ç†å‘˜æƒé™ä»¥è®¿é—®åº•å±‚å­˜å‚¨è®¾å¤‡ä¸Žç³»ç»Ÿèµ„æºã€‚")
		}
	}

	// Automatically notify frontend when removable USB storage devices are plugged or unplugged
	disk.StartHotplugMonitor(a.ctx, func() {
		logger.Debug("Removable storage hotplug change detected by background monitor")
		a.emitEvent("disk:hotplug")
		disk.InvalidateDiskCache()
	})

	hypervisor.RegisterVMExitHandler(func(targetDisk string, vmErr error) {
		if vmErr != nil {
			logger.Warn("Virtual machine session terminated with error", "disk", targetDisk, "error", vmErr)
		} else {
			logger.Info("Virtual machine session exited normally", "disk", targetDisk)
		}
		disk.InvalidateDiskCache()
		a.emitEvent("vm:exit", map[string]interface{}{
			"disk": targetDisk,
			"error": func() string {
				if vmErr != nil {
					return vmErr.Error()
				}
				return ""
			}(),
		})
	})

	logger.Info("UniGoDesktop Wails GUI runtime started successfully")
}

// shutdown is called when the Wails application is shutting down.
func (a *App) shutdown(ctx context.Context) {
	logger.SetWailsContext(nil)
	if a.cancel != nil {
		a.cancel()
	}

	hypervisor.GetManager().CleanupAllUnmountedDisks()

	// Cleanly disconnect and terminate active privileged worker session
	if client := privilege.GetActiveWorkerClient(); client != nil {
		_ = client.Terminate()
		privilege.SetActiveWorkerClient(nil)
		if runtime.GOOS == "windows" {
			time.Sleep(150 * time.Millisecond)
		}
	}

	logger.Info("UniGoDesktop Wails GUI runtime shutting down")
}

// GetRecentLogs retrieves the in-memory buffered logs.
func (a *App) GetRecentLogs() []logger.LogEntry {
	return logger.GetRecentLogs()
}

// ClearLogs clears the in-memory buffered logs.
func (a *App) ClearLogs() {
	logger.ClearLogs()
}

// LogAction records structured user or frontend interactions.
func (a *App) LogAction(level, message, details string) {
	switch strings.ToUpper(level) {
	case "ERROR":
		logger.Error(message, "details", details)
	case "WARN":
		logger.Warn(message, "details", details)
	case "DEBUG":
		logger.Debug(message, "details", details)
	default:
		logger.Info(message, "details", details)
	}
}

// ExportLogs opens a native save file dialog to export log content to a file (.log or .txt).
func (a *App) ExportLogs(content, title, logFilter, textFilter, allFilter string) (string, error) {
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
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		logger.Error("Failed to write log export file", "path", filePath, "error", err)
		return "", fmt.Errorf("write log file: %w", err)
	}
	logger.Info("Logs exported successfully", "path", filePath)
	return filePath, nil
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
		return true // prevent application exit
	}

	logger.Info("Window close proceeding: quitting application")
	return false // allow application exit
}

// onSecondInstanceLaunch is invoked when a second instance of the application attempts to start.
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	logger.Info(fmt.Sprintf("Second instance launch detected with args: %v", secondInstanceData.Args))
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	}
}

// Greet returns a friendly greeting for demonstration.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello %s, Welcome to UniGoDesktop!", name)
}

// GetHelloInfo returns structured hello greeting and runtime environment details.
func (a *App) GetHelloInfo() *HelloInfo {
	return &HelloInfo{
		Greeting:  "Hello World From UniGoDesktop!",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// GetSystemInfo returns system environment and runtime diagnostic metadata.
func (a *App) GetSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		AppName:   "UniGoDesktop",
		Version:   env.GitTag,
		Commit:    env.CommitHash,
		BuildTime: env.BuildTime,
		DataDir:   env.GetDataDir(),
		ConfigDir: env.GetConfigDir(),
	}
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

// CheckUpdate returns GitHub release update metadata.
func (a *App) CheckUpdate() *updater.UpdateStatus {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return updater.CheckUpdate(ctx)
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
		a.emitEvent("gui-update-progress", p)
	}

	return updater.PerformGuiUpdate(ctx, proxyPrefix, progressCallback)
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
	if cfg.ProxyPassword != "" {
		if err := config.SaveProxyPassword(cfg.ProxyPassword); err != nil {
			logger.Error("Failed to save proxy password to system credential store", "error", err)
			return err
		}
		cfg.ProxyPassword = ""
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	a.SetTheme(cfg.Theme)
	return nil
}

// SetTheme updates the application window appearance and synchronizes the native titlebar theme.
func (a *App) SetTheme(theme string) {
	if a.ctx != nil && a.ctx.Value("frontend") != nil {
		switch theme {
		case "dark":
			wailsRuntime.WindowSetDarkTheme(a.ctx)
			wailsRuntime.WindowSetBackgroundColour(a.ctx, 7, 10, 18, 255)
		case "light":
			wailsRuntime.WindowSetLightTheme(a.ctx)
			wailsRuntime.WindowSetBackgroundColour(a.ctx, 241, 245, 249, 255)
		default:
			wailsRuntime.WindowSetSystemDefaultTheme(a.ctx)
			if sysinfo.IsSystemDarkTheme() {
				wailsRuntime.WindowSetBackgroundColour(a.ctx, 7, 10, 18, 255)
			} else {
				wailsRuntime.WindowSetBackgroundColour(a.ctx, 241, 245, 249, 255)
			}
		}
	}
	sysinfo.SyncTitleBarTheme(theme)
	logger.Info("Synchronized application and titlebar theme", "theme", theme)
}

// ClearProxyPassword removes the saved proxy password from the system credential store.
func (a *App) ClearProxyPassword() error {
	return config.DeleteProxyPassword()
}

// OpenURL opens the specified URL in the native desktop browser.
func (a *App) OpenURL(url string) {
	if url != "" && a.ctx != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, url)
	}
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
				if a.ctx != nil {
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

	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
	return nil
}

// ReloadAppMenu rebuilds and updates the native application menu with the specified language.
func (a *App) ReloadAppMenu(lang string) error {
	if strings.HasPrefix(strings.ToLower(lang), "zh") {
		privilege.SetDefaultElevationPrompt("UniGoDesktop éœ€è¦ç®¡ç†å‘˜æƒé™ä»¥è®¿é—®åº•å±‚å­˜å‚¨è®¾å¤‡ä¸Žç³»ç»Ÿèµ„æºã€‚")
	} else {
		privilege.SetDefaultElevationPrompt("UniGoDesktop requires administrator privileges to access raw storage devices and system resources.")
	}

	if a.ctx == nil {
		return nil
	}
	appMenu := BuildAppMenu(a, lang)
	wailsRuntime.MenuSetApplicationMenu(a.ctx, appMenu)
	wailsRuntime.MenuUpdateApplicationMenu(a.ctx)
	return nil
}

// GetDiskList returns all removable storage devices safely filtered.
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

// DetectHypervisors returns status of all installed virtual machine engines.
func (a *App) DetectHypervisors() []*hypervisor.VMStatus {
	return hypervisor.GetManager().DetectAll()
}

// DetectBestHypervisor returns the highest priority available virtual machine status.
func (a *App) DetectBestHypervisor() *hypervisor.VMStatus {
	return hypervisor.GetManager().DetectBest()
}

// GetDefaultVMConfig returns recommended default VM tuning parameters.
func (a *App) GetDefaultVMConfig() *hypervisor.VMConfig {
	return hypervisor.DefaultVMConfig()
}

// LaunchVM launches a specified or best available virtual machine with boot mode (uefi, bios, auto).
func (a *App) LaunchVM(targetDisk, vmType, bootMode string) error {
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

// LaunchVMWithConfig launches a virtual machine with custom VMConfig options.
func (a *App) LaunchVMWithConfig(targetDisk, vmType string, cfg hypervisor.VMConfig) error {
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
	a.emitEvent("disk:hotplug")
	return nil
}

// StopVM manually stops any active virtual machine simulation session, remounts target disks, and resets state.
func (a *App) StopVM() error {
	logger.Info("User manually requested VM simulation session stop")
	hypervisor.StopActiveVMSession()
	hypervisor.GetManager().CleanupAllUnmountedDisks()
	a.emitEvent("vm:exit", map[string]interface{}{"disk": "", "error": ""})
	return nil
}

// IsPrivileged returns true if the app process or worker currently possesses administrator or root privileges.
func (a *App) IsPrivileged() bool {
	return privilege.IsElevated()
}

// RequestPrivilegeElevation prompts the user for administrator privileges across operating systems.
// An optional custom prompt string can be provided to support localized authorization dialogs.
func (a *App) RequestPrivilegeElevation(customPrompt string) (bool, error) {
	if privilege.IsElevated() {
		return true, nil
	}

	prompt := strings.TrimSpace(customPrompt)
	if prompt != "" {
		privilege.SetDefaultElevationPrompt(prompt)
	} else {
		cfg, _ := config.Load()
		lang := ""
		if cfg != nil {
			lang = strings.ToLower(cfg.Language)
		}
		if strings.HasPrefix(lang, "zh") {
			prompt = "UniGoDesktop éœ€è¦ç®¡ç†å‘˜æƒé™ä»¥è®¿é—®åº•å±‚å­˜å‚¨è®¾å¤‡ä¸Žç³»ç»Ÿèµ„æºã€‚"
		} else {
			prompt = privilege.GetDefaultElevationPrompt()
		}
	}

	_, err := privilege.StartOrConnectWorker(prompt)
	if err != nil {
		if strings.Contains(err.Error(), "canceled") || strings.Contains(err.Error(), "rejected") {
			logger.Info("User dismissed privilege elevation prompt")
			return false, nil
		}
		logger.Warn("Failed to start privileged worker", "error", err)
		return false, err
	}

	privilege.ResetElevationCache()
	disk.InvalidateDiskCache()
	logger.Info("Administrator privilege successfully granted by user")
	return true, nil
}

// CheckConfigHealth performs an integrity diagnosis on the config file.
func (a *App) CheckConfigHealth() (*config.ConfigHealth, error) {
	return config.HealthCheck()
}

// CalculateFileChecksum calculates the hash (sha256, md5, sha1, crc32, etc.) of a local file.
func (a *App) CalculateFileChecksum(filePath, algo string) (*utils.ChecksumResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return utils.CalculateFileChecksum(ctx, filePath, algo)
}
