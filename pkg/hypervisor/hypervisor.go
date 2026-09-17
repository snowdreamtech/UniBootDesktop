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

// Driver defines the standard interface for hypervisor implementations.
type Driver interface {
	Type() HypervisorType
	Name() string
	Priority() int
	Detect() *VMStatus
	Launch(ctx context.Context, targetDisk string, bootMode string) error
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
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, drv := range m.drivers {
		status := drv.Detect()
		if status != nil && status.Installed {
			logger.Info("Selected best available hypervisor for preview launch", "hypervisor", drv.Name(), "disk", targetDisk, "bootMode", bootMode)
			return drv.Launch(ctx, targetDisk, bootMode)
		}
	}

	// Fallback check for dry-run
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		logger.Info("Dry-run hypervisor simulation test executed", "disk", targetDisk, "bootMode", bootMode)
		return nil
	}

	return fmt.Errorf("no supported virtual machine (QEMU, UTM, VMware, VirtualBox) detected on host system")
}

// LaunchSpecified launches a specific hypervisor driver by type.
func (m *Manager) LaunchSpecified(ctx context.Context, targetDisk string, hType HypervisorType, bootMode string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, drv := range m.drivers {
		if drv.Type() == hType {
			status := drv.Detect()
			if status == nil || !status.Installed {
				return fmt.Errorf("requested hypervisor '%s' is not installed", drv.Name())
			}
			logger.Info("Launching specified hypervisor", "hypervisor", drv.Name(), "disk", targetDisk, "bootMode", bootMode)
			return drv.Launch(ctx, targetDisk, bootMode)
		}
	}

	return fmt.Errorf("unknown hypervisor type '%s'", hType)
}

// unmountTargetDisk safely unmounts disk partitions across macOS, Linux, and Windows before hypervisor launch.
func unmountTargetDisk(targetPath string) {
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		return
	}
	logger.Info("Safely unmounting target disk partitions before VM launch", "targetPath", targetPath)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		cmd := exec.Command("diskutil", "unmountDisk", "force", fmt.Sprintf("/dev/%s", diskNode))
		_ = cmd.Run()

		// Poll up to 2.5s for macOS kernel/diskarbitrationd lock release to prevent "Resource busy"
		deadline := time.Now().Add(2500 * time.Millisecond)
		for time.Now().Before(deadline) {
			time.Sleep(150 * time.Millisecond)
			if f, err := os.OpenFile(targetPath, os.O_RDWR, 0); err == nil {
				_ = f.Close()
				break
			} else if os.IsPermission(err) {
				break
			} else if strings.Contains(err.Error(), "busy") {
				_ = exec.Command("diskutil", "unmountDisk", "force", fmt.Sprintf("/dev/%s", diskNode)).Run()
			}
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("udisksctl", "unmount", "-b", targetPath)
		if err := cmd.Run(); err != nil {
			_ = exec.Command("umount", targetPath).Run()
		}
		time.Sleep(300 * time.Millisecond)
	} else if runtime.GOOS == "windows" {
		psCmd := fmt.Sprintf(`Get-Volume | Where-DriveLetter | Where-Object { $_.Path -like "*%s*" } | Dismount-Volume -Confirm:$false`, targetPath)
		_ = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).Run()
		time.Sleep(300 * time.Millisecond)
	}
}

// remountTargetDisk automatically remounts target disk partitions back to host OS after VM exit.
func remountTargetDisk(targetPath string) {
	if targetPath == "" || os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		return
	}
	logger.Info("Remounting target disk partitions back to host OS after VM exit", "targetPath", targetPath)

	if runtime.GOOS == "darwin" {
		diskNode := strings.TrimPrefix(targetPath, "/dev/rdisk")
		diskNode = strings.TrimPrefix(diskNode, "/dev/disk")
		if !strings.HasPrefix(diskNode, "disk") {
			diskNode = "disk" + diskNode
		}
		cmd := exec.Command("diskutil", "mountDisk", fmt.Sprintf("/dev/%s", diskNode))
		_ = cmd.Run()
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("udisksctl", "mount", "-b", targetPath)
		_ = cmd.Run()
	} else if runtime.GOOS == "windows" {
		psCmd := fmt.Sprintf(`Get-Volume | Where-DriveLetter | Where-Object { $_.Path -like "*%s*" } | Mount-Volume`, targetPath)
		_ = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).Run()
	}
}
