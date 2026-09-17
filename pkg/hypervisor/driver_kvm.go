// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

type KVMDriver struct{}

func (d *KVMDriver) Type() HypervisorType {
	return TypeKVM
}

func (d *KVMDriver) Name() string {
	return "KVM / Virt-Manager / GNOME Boxes"
}

func (d *KVMDriver) Priority() int {
	return 2
}

func (d *KVMDriver) Detect() *VMStatus {
	if runtime.GOOS != "linux" {
		return &VMStatus{
			Type:       TypeKVM,
			Name:       d.Name(),
			Installed:  false,
			Path:       "",
			Version:    "Supported only on Linux OS",
			Priority:   d.Priority(),
			CanBootRaw: false,
		}
	}

	// 1. Check gnome-boxes
	if path, err := exec.LookPath("gnome-boxes"); err == nil {
		return &VMStatus{
			Type:       TypeKVM,
			Name:       "GNOME Boxes (KVM)",
			Installed:  true,
			Path:       path,
			Version:    "GNOME Boxes (KVM)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Check virt-manager
	if path, err := exec.LookPath("virt-manager"); err == nil {
		return &VMStatus{
			Type:       TypeKVM,
			Name:       "Virt-Manager (KVM/libvirt)",
			Installed:  true,
			Path:       path,
			Version:    "Virt-Manager (KVM/libvirt)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 3. Check kvm command / /dev/kvm device
	if path, err := exec.LookPath("kvm"); err == nil {
		return &VMStatus{
			Type:       TypeKVM,
			Name:       "Kernel-based Virtual Machine (KVM)",
			Installed:  true,
			Path:       path,
			Version:    "KVM Hypervisor",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	if _, err := os.Stat("/dev/kvm"); err == nil {
		return &VMStatus{
			Type:       TypeKVM,
			Name:       "Kernel-based Virtual Machine (/dev/kvm)",
			Installed:  true,
			Path:       "/dev/kvm",
			Version:    "Linux KVM Kernel Module",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	return &VMStatus{
		Type:       TypeKVM,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *KVMDriver) Launch(ctx context.Context, diskPath string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed or enabled on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run KVM launch complete", "diskPath", diskPath)
		return nil
	}

	logger.Info("Executing KVM preview simulation test", "disk", diskPath, "path", status.Path)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	ensureDiskPermissions(targetPath)

	cmd := exec.Command(status.Path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch KVM tool: %w", err)
	}

	return nil
}
