// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"sync"

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

// Driver defines the standard interface for hypervisor implementations.
type Driver interface {
	Type() HypervisorType
	Name() string
	Priority() int
	Detect() *VMStatus
	Launch(ctx context.Context, targetDisk string) error
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
func (m *Manager) LaunchBest(ctx context.Context, targetDisk string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, drv := range m.drivers {
		status := drv.Detect()
		if status != nil && status.Installed {
			logger.Info("Selected best available hypervisor for preview launch", "hypervisor", drv.Name(), "disk", targetDisk)
			return drv.Launch(ctx, targetDisk)
		}
	}

	// Fallback check for dry-run
	if os.Getenv("UNIBOOT_DRY_RUN") != "" {
		logger.Info("Dry-run hypervisor simulation test executed", "disk", targetDisk)
		return nil
	}

	return fmt.Errorf("no supported virtual machine (QEMU, UTM, VMware, VirtualBox) detected on host system")
}

// LaunchSpecified launches a specific hypervisor driver by type.
func (m *Manager) LaunchSpecified(ctx context.Context, targetDisk string, hType HypervisorType) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, drv := range m.drivers {
		if drv.Type() == hType {
			status := drv.Detect()
			if status == nil || !status.Installed {
				return fmt.Errorf("requested hypervisor '%s' is not installed", drv.Name())
			}
			logger.Info("Launching specified hypervisor", "hypervisor", drv.Name(), "disk", targetDisk)
			return drv.Launch(ctx, targetDisk)
		}
	}

	return fmt.Errorf("unknown hypervisor type '%s'", hType)
}
