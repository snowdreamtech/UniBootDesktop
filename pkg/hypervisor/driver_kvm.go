// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

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

func (d *KVMDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	status := d.Detect()
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run KVM launch complete", "diskPath", diskPath, "bootMode", bootMode)
		return nil
	}

	logger.Info("Executing Linux KVM / Virt-Manager preview test instance", "disk", diskPath, "kvmPath", status.Path, "bootMode", bootMode)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	restoreDiskPerms, err := ensureDiskPermissions(targetPath)
	if err != nil {
		remountTargetDisk(targetPath)
		return fmt.Errorf("failed to acquire target disk permissions: %w", err)
	}
	unmountTargetDisk(targetPath)

	cmd := exec.Command(status.Path)
	if err := cmd.Start(); err != nil {
		remountTargetDisk(targetPath)
		restoreDiskPerms()
		return fmt.Errorf("failed to launch KVM tool: %w", err)
	}

	go func() {
		err := cmd.Wait()
		time.Sleep(300 * time.Millisecond)
		remountTargetDisk(targetPath)
		restoreDiskPerms()
		NotifyVMExited(targetPath, err)
	}()

	return nil
}

// LaunchWithConfig launches KVM with custom VMConfig options.
func (d *KVMDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	return d.Launch(ctx, diskPath, cfg.BootMode)
}

