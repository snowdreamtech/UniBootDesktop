// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
	"github.com/snowdreamtech/unigodesktop/pkg/privilege"
)

// HypervisorType defines supported virtualization engines.
type HypervisorType string

const (
	TypeQEMU       HypervisorType = "qemu"
	TypeUTM        HypervisorType = "utm"
	TypeKVM        HypervisorType = "kvm"
	TypeParallels  HypervisorType = "parallels"
	TypeVMware     HypervisorType = "vmware"
	TypeHyperV     HypervisorType = "hyperv"
	TypeVirtualBox HypervisorType = "virtualbox"
)

// VMStatus contains metadata about an installed hypervisor.
type VMStatus struct {
	Type       HypervisorType `json:"type"`
	Name       string         `json:"name"`
	Installed  bool           `json:"installed"`
	Path       string         `json:"path"`
	Version    string         `json:"version"`
	Priority   int            `json:"priority"`
	CanBootRaw bool           `json:"canBootRaw"`
}

const (
	BootModeAuto = "auto"
	BootModeUEFI = "uefi"
	BootModeBIOS = "bios"
)

// temporaryDiskAccessWindow is how long a fire-and-forget GUI hypervisor may
// keep the owner-only temporary device mode before permissions are restored.
const temporaryDiskAccessWindow = 2 * time.Second

// scheduleDiskPermissionRestore restores raw-disk modes either when waitFn
// returns (preferred) or after temporaryDiskAccessWindow if waitFn is nil.
func scheduleDiskPermissionRestore(restore func(), waitFn func() error) {
	if restore == nil {
		return
	}
	go func() {
		if waitFn != nil {
			_ = waitFn()
		} else {
			time.Sleep(temporaryDiskAccessWindow)
		}
		restore()
	}()
}

// VMConfig contains customizable hardware and virtual machine simulation options.
type VMConfig struct {
	CpuCores     int    `json:"cpuCores"`     // 1, 2, 4, 8 cores (Default: 2)
	MemoryMB     int    `json:"memoryMB"`     // 1024, 2048, 4096, 8192 MB (Default: 2048)
	BootMode     string `json:"bootMode"`     // "auto", "uefi", "bios"
	DisplayAccel bool   `json:"displayAccel"` // Enable hardware acceleration (-accel hvf/kvm/haxm)
	SecureBoot   bool   `json:"secureBoot"`   // SecureBoot OVMF simulation
}

// DefaultVMConfig returns standard recommended virtual machine configuration settings.
func DefaultVMConfig() *VMConfig {
	return &VMConfig{
		CpuCores:     2,
		MemoryMB:     2048,
		BootMode:     BootModeAuto,
		DisplayAccel: true,
		SecureBoot:   false,
	}
}

// Driver defines the standard interface for hypervisor implementations.
type Driver interface {
	Type() HypervisorType
	Name() string
	Priority() int
	Detect() *VMStatus
	Launch(ctx context.Context, targetDisk string, bootMode string) error
}

// ConfigurableDriver extends Driver interface to accept custom VMConfig settings.
type ConfigurableDriver interface {
	Driver
	LaunchWithConfig(ctx context.Context, targetDisk string, cfg VMConfig) error
}

// Manager orchestrates hypervisor detection and priority fallback.
type Manager struct {
	drivers []Driver
	mu      sync.RWMutex
}

var (
	defaultManager *Manager
	once           sync.Once
)

// GetManager returns the global hypervisor manager singleton.
func GetManager() *Manager {
	once.Do(func() {
		defaultManager = &Manager{
			drivers: make([]Driver, 0),
		}
		// Register default drivers in order of priority: QEMU -> UTM -> KVM -> Parallels -> VMware -> Hyper-V -> VirtualBox
		defaultManager.Register(&QEMUDriver{})
		defaultManager.Register(&UTMDriver{})
		defaultManager.Register(&KVMDriver{})
		defaultManager.Register(&ParallelsDriver{})
		defaultManager.Register(&VMwareDriver{})
		defaultManager.Register(&HyperVDriver{})
		defaultManager.Register(&VirtualBoxDriver{})
	})
	return defaultManager
}

// Register adds a new hypervisor driver to the manager.
func (m *Manager) Register(driver Driver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers = append(m.drivers, driver)
}

// DetectAll scans system for all registered hypervisors and returns their statuses.
func (m *Manager) DetectAll() []*VMStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*VMStatus, 0, len(m.drivers))
	for _, drv := range m.drivers {
		status := drv.Detect()
		results = append(results, status)
	}
	return results
}

// DetectBest returns the highest priority available hypervisor.
func (m *Manager) DetectBest() *VMStatus {
	statuses := m.DetectAll()
	for _, st := range statuses {
		if st != nil && st.Installed {
			return st
		}
	}

	// Dry-run mode for tests or simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		return &VMStatus{
			Type:       TypeQEMU,
			Name:       "QEMU Simulator (Mock)",
			Installed:  true,
			Path:       "/usr/local/bin/qemu-system-x86_64 (Dry-Run)",
			Version:    "QEMU 8.2 (Dry-Run)",
			Priority:   1,
			CanBootRaw: true,
		}
	}

	return &VMStatus{
		Type:       TypeQEMU,
		Name:       "QEMU",
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   1,
		CanBootRaw: false,
	}
}

// LaunchBest launches the first available hypervisor according to priority chain.
func (m *Manager) LaunchBest(ctx context.Context, targetDisk string, bootMode string) error {
	return m.LaunchBestConfigured(ctx, targetDisk, VMConfig{
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		BootMode:     bootMode,
		DisplayAccel: true,
		SecureBoot:   false,
	})
}

// LaunchBestConfigured launches the first available hypervisor with custom VMConfig options.
func (m *Manager) LaunchBestConfigured(ctx context.Context, targetDisk string, cfg VMConfig) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	isRawDisk := strings.HasPrefix(targetDisk, "/dev/") || strings.HasPrefix(targetDisk, `\\.\PhysicalDrive`) || strings.HasPrefix(targetDisk, "PhysicalDrive")
	for _, drv := range m.drivers {
		status := drv.Detect()
		if status != nil && status.Installed {
			if isRawDisk && !status.CanBootRaw {
				continue
			}
			logger.Info("Selected best available hypervisor for preview launch", "hypervisor", drv.Name(), "disk", targetDisk, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)
			if cDrv, ok := drv.(ConfigurableDriver); ok {
				return cDrv.LaunchWithConfig(ctx, targetDisk, cfg)
			}
			return drv.Launch(ctx, targetDisk, cfg.BootMode)
		}
	}

	// Fallback check for dry-run
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		logger.Info("Dry-run hypervisor simulation test executed", "disk", targetDisk, "bootMode", cfg.BootMode)
		return nil
	}

	return fmt.Errorf("no supported virtual machine (QEMU, UTM, VMware, VirtualBox) detected on host system")
}

// LaunchSpecified launches a specific hypervisor driver by type.
func (m *Manager) LaunchSpecified(ctx context.Context, targetDisk string, hType HypervisorType, bootMode string) error {
	return m.LaunchSpecifiedConfigured(ctx, targetDisk, hType, VMConfig{
		CpuCores:     GetRecommendedVCPUs(),
		MemoryMB:     GetRecommendedVMMemoryMB(),
		BootMode:     bootMode,
		DisplayAccel: true,
		SecureBoot:   false,
	})
}

// LaunchSpecifiedConfigured launches a specific hypervisor driver with custom VMConfig options.
func (m *Manager) LaunchSpecifiedConfigured(ctx context.Context, targetDisk string, hType HypervisorType, cfg VMConfig) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, drv := range m.drivers {
		if drv.Type() == hType {
			status := drv.Detect()
			if (status == nil || !status.Installed) && os.Getenv("UNIBOOT_DRY_RUN") == "" {
				return fmt.Errorf("requested hypervisor '%s' is not installed", drv.Name())
			}
			logger.Info("Launching specified hypervisor", "hypervisor", drv.Name(), "disk", targetDisk, "bootMode", cfg.BootMode, "cpu", cfg.CpuCores, "ramMB", cfg.MemoryMB)
			if cDrv, ok := drv.(ConfigurableDriver); ok {
				return cDrv.LaunchWithConfig(ctx, targetDisk, cfg)
			}
			return drv.Launch(ctx, targetDisk, cfg.BootMode)
		}
	}

	return fmt.Errorf("unknown hypervisor type '%s'", hType)
}

var unmountedDisksTracker sync.Map

// TrackDiskUnmounted registers a disk path as currently unmounted for VM preview.
func TrackDiskUnmounted(targetPath string) {
	if targetPath != "" {
		unmountedDisksTracker.Store(targetPath, true)
	}
}

// TrackDiskRemounted unregisters a disk path after VM exit remount.
func TrackDiskRemounted(targetPath string) {
	if targetPath != "" {
		unmountedDisksTracker.Delete(targetPath)
	}
}

// CleanupAllUnmountedDisks remounts only the disks previously tracked by the VM lifecycle.
// This intentionally does NOT enumerate all removable media or eject arbitrary USB devices.
// The shutdown policy is: best-effort remount of tracked VM targets, never mass-eject all U disks.
func (m *Manager) CleanupAllUnmountedDisks() {
	var paths []string
	unmountedDisksTracker.Range(func(key, value any) bool {
		if path, ok := key.(string); ok && path != "" {
			paths = append(paths, path)
		}
		return true
	})

	for _, path := range paths {
		logger.Info("Emergency cleanup: remounting tracked target disk back to host OS", "targetPath", path)
		remountTargetDisk(path)
		unmountedDisksTracker.Delete(path)
	}
}

func runCommandWithTimeout(timeout time.Duration, name string, args ...string) error {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	if err := privilege.ValidateCommandName(name); err != nil {
		return fmt.Errorf("unsafe hypervisor command: %w", err)
	}
	for _, arg := range args {
		if err := privilege.ValidateCommandArgument(arg); err != nil {
			return fmt.Errorf("unsafe hypervisor argument: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %s: %w", timeout, ctx.Err())
		}
		return err
	}
	return nil
}

// unmountTargetDisk safely unmounts disk partitions across macOS, Linux, and Windows before hypervisor launch.
func unmountTargetDisk(targetPath string) {
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		return
	}
	TrackDiskUnmounted(targetPath)
	logger.Info("Safely unmounting target disk partitions before VM launch", "targetPath", targetPath)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		cmdPath := fmt.Sprintf("/dev/%s", diskNode)
		_ = runCommandWithTimeout(3*time.Second, "diskutil", "unmountDisk", "force", cmdPath)

		deadline := time.Now().Add(3000 * time.Millisecond)
		for time.Now().Before(deadline) {
			if !isDiskMounted(diskNode) {
				break
			}
			time.Sleep(150 * time.Millisecond)
			_ = runCommandWithTimeout(3*time.Second, "diskutil", "unmountDisk", "force", cmdPath)
		}
	} else if runtime.GOOS == "linux" {
		if err := runCommandWithTimeout(3*time.Second, "udisksctl", "unmount", "-b", targetPath); err != nil {
			_ = runCommandWithTimeout(3*time.Second, "umount", targetPath)
		}
		time.Sleep(300 * time.Millisecond)
	} else if runtime.GOOS == "windows" {
		// 安全地转义PowerShell参数，防止命令注入
		// 移除可能的PowerShell注入字符
		safePath := strings.ReplaceAll(targetPath, "'", "''") // PowerShell单引号转义
		safePath = strings.ReplaceAll(safePath, "`", "``")    // PowerShell反引号转义
		safePath = strings.ReplaceAll(safePath, "$", "`$")    // PowerShell变量转义
		safePath = strings.ReplaceAll(safePath, "\"", "`\"")  // 双引号转义

		psCmd := fmt.Sprintf(`Get-Volume | Where-DriveLetter | Where-Object { $_.Path -like '*%s*' } | Dismount-Volume -Confirm:$false`, safePath)
		_ = runCommandWithTimeout(3*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
		time.Sleep(300 * time.Millisecond)
	}
}

// remountTargetDisk automatically remounts target disk partitions back to host OS after VM exit.
func remountTargetDisk(targetPath string) {
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		return
	}
	defer TrackDiskRemounted(targetPath)
	logger.Info("Remounting target disk partitions back to host OS after VM exit", "targetPath", targetPath)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		_ = runCommandWithTimeout(3*time.Second, "diskutil", "mountDisk", fmt.Sprintf("/dev/%s", diskNode))
	} else if runtime.GOOS == "linux" {
		_ = runCommandWithTimeout(3*time.Second, "udisksctl", "mount", "-b", targetPath)
	} else if runtime.GOOS == "windows" {
		// 安全地转义PowerShell参数，防止命令注入
		safePath := strings.ReplaceAll(targetPath, "'", "''") // PowerShell单引号转义
		safePath = strings.ReplaceAll(safePath, "`", "``")    // PowerShell反引号转义
		safePath = strings.ReplaceAll(safePath, "$", "`$")    // PowerShell变量转义
		safePath = strings.ReplaceAll(safePath, "\"", "`\"")  // 双引号转义

		psCmd := fmt.Sprintf(`Get-Volume | Where-DriveLetter | Where-Object { $_.Path -like '*%s*' } | Mount-Volume`, safePath)
		_ = runCommandWithTimeout(3*time.Second, "powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	}
}

// isDiskMounted checks if a macOS disk or any of its partitions are currently mounted.
func isDiskMounted(diskNode string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	nodesToTest := []string{diskNode}
	for i := 1; i <= 8; i++ {
		nodesToTest = append(nodesToTest, fmt.Sprintf("%ss%d", diskNode, i))
	}
	for _, node := range nodesToTest {
		cmd := exec.Command("diskutil", "info", fmt.Sprintf("/dev/%s", node))
		out, err := cmd.Output()
		if err == nil {
			str := string(out)
			if strings.Contains(str, "Mounted:                   Yes") || strings.Contains(str, "Mounted: Yes") || strings.Contains(str, "Mount Point:") {
				return true
			}
		}
	}
	return false
}

// GetRecommendedVCPUs dynamically calculates optimal VM CPU core count as roughly ~1/4 of total host CPU cores,
// with safe bounds (min 1 vCPU, max 4 vCPUs).
func GetRecommendedVCPUs() int {
	cpus := runtime.NumCPU()
	vcpus := cpus / 4
	if vcpus < 2 {
		vcpus = 2
	}
	if cpus <= 2 {
		vcpus = 1
	}
	if vcpus > 4 {
		vcpus = 4
	}
	return vcpus
}

// GetRecommendedVMMemoryMB dynamically calculates optimal VM RAM (in MB)
// proportional to host RAM (~1/4 total RAM) with bounds (min 2048MB, max 8192MB),
// allocating up to 8GB RAM on 24GB+ host systems while protecting 4GB host RAM machines.
func GetRecommendedVMMemoryMB() int {
	totalRAMBytes := getHostTotalRAMBytes()
	if totalRAMBytes == 0 {
		return 2048 // Safe default fallback
	}

	totalMB := int(totalRAMBytes / (1024 * 1024))
	recommendedMB := totalMB / 4

	if recommendedMB < 2048 {
		recommendedMB = 2048
	}
	if recommendedMB > 8192 {
		recommendedMB = 8192
	}

	// Safety cap for <= 4.5GB host RAM machines to prevent host OS thrashing
	if totalMB <= 4608 && recommendedMB > 2048 {
		recommendedMB = 2048
	}

	logger.Info("Proportional VM RAM recommendation evaluated", "hostRAMMB", totalMB, "recommendedRAMMB", recommendedMB)
	return recommendedMB
}

func getHostTotalRAMBytes() uint64 {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			var bytes uint64
			if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &bytes); err == nil {
				return bytes
			}
		}
	case "linux":
		data, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						var kb uint64
						if _, err := fmt.Sscanf(fields[1], "%d", &kb); err == nil {
							return kb * 1024
						}
					}
				}
			}
		}
	case "windows":
		out, err := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, "TotalPhysicalMemory") {
					var bytes uint64
					if _, err := fmt.Sscanf(line, "%d", &bytes); err == nil {
						return bytes
					}
				}
			}
		}
	}
	return 0
}
