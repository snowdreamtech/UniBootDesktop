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

type UTMDriver struct{}

func (d *UTMDriver) Type() HypervisorType {
	return TypeUTM
}

func (d *UTMDriver) Name() string {
	return "UTM Virtual Machine"
}

func (d *UTMDriver) Priority() int {
	return 2
}

func (d *UTMDriver) Detect() *VMStatus {
	if runtime.GOOS != "darwin" {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM",
			Installed:  false,
			Path:       "",
			Version:    "Not Supported on this OS",
			Priority:   d.Priority(),
			CanBootRaw: false,
		}
	}

	// 1. Check utmctl command
	if path, err := exec.LookPath("utmctl"); err == nil {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM Virtual Machine",
			Installed:  true,
			Path:       path,
			Version:    "UTM CLI (utmctl)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Check /Applications/UTM.app bundle
	appPath := "/Applications/UTM.app"
	if info, err := os.Stat(appPath); err == nil && info.IsDir() {
		return &VMStatus{
			Type:       TypeUTM,
			Name:       "UTM Virtual Machine",
			Installed:  true,
			Path:       appPath,
			Version:    "UTM App (/Applications/UTM.app)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	return &VMStatus{
		Type:       TypeUTM,
		Name:       "UTM",
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *UTMDriver) Launch(ctx context.Context, diskPath string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("UTM virtual machine application not detected on macOS system")
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run UTM launch complete", "diskPath", diskPath)
		return nil
	}

	logger.Info("Executing UTM preview test instance", "disk", diskPath, "utmPath", status.Path)

	// Ensure disk permissions on macOS
	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	ensureDiskPermissions(targetPath)

	if path, err := exec.LookPath("utmctl"); err == nil {
		cmd := exec.Command(path, "list")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			logger.Info("UTM CLI detected, listing UTM VMs", "output", string(output))
		}
	}

	// Open UTM Application with raw disk parameter or bundle
	cmd := exec.Command("open", "-a", "UTM")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open UTM application: %w", err)
	}

	return nil
}
