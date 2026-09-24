// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hypervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

type HyperVDriver struct{}

func (d *HyperVDriver) Type() HypervisorType {
	return TypeHyperV
}

func (d *HyperVDriver) Name() string {
	return "Microsoft Hyper-V"
}

func (d *HyperVDriver) Priority() int {
	return 5
}

func (d *HyperVDriver) Detect() *VMStatus {
	if runtime.GOOS != "windows" {
		return &VMStatus{
			Type:       TypeHyperV,
			Name:       d.Name(),
			Installed:  false,
			Path:       "",
			Version:    "Supported only on Windows OS",
			Priority:   d.Priority(),
			CanBootRaw: false,
		}
	}

	// 1. Check vmconnect.exe in System32
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	vmconnectPath := filepath.Join(windir, "System32", "vmconnect.exe")
	if info, err := os.Stat(vmconnectPath); err == nil && !info.IsDir() {
		return &VMStatus{
			Type:       TypeHyperV,
			Name:       d.Name(),
			Installed:  true,
			Path:       vmconnectPath,
			Version:    "Hyper-V Manager (vmconnect.exe)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Check vmconnect in PATH
	if path, err := exec.LookPath("vmconnect.exe"); err == nil {
		return &VMStatus{
			Type:       TypeHyperV,
			Name:       d.Name(),
			Installed:  true,
			Path:       path,
			Version:    "Hyper-V Manager (vmconnect)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	return &VMStatus{
		Type:       TypeHyperV,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed / Enabled",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *HyperVDriver) Launch(ctx context.Context, diskPath string, bootMode string) error {
	status := d.Detect()
	if !status.Installed && os.Getenv("UNIBOOT_DRY_RUN") == "" {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run Hyper-V launch complete", "diskPath", diskPath, "bootMode", bootMode)
		return nil
	}

	logger.Info("Executing Hyper-V preview test instance", "disk", diskPath, "hypervPath", status.Path, "bootMode", bootMode)

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
		return fmt.Errorf("failed to launch Hyper-V connection tool: %w", err)
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

// LaunchWithConfig launches Hyper-V with custom VMConfig options.
func (d *HyperVDriver) LaunchWithConfig(ctx context.Context, diskPath string, cfg VMConfig) error {
	return d.Launch(ctx, diskPath, cfg.BootMode)
}

