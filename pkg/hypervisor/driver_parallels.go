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

type ParallelsDriver struct{}

func (d *ParallelsDriver) Type() HypervisorType {
	return TypeParallels
}

func (d *ParallelsDriver) Name() string {
	return "Parallels Desktop"
}

func (d *ParallelsDriver) Priority() int {
	return 3
}

func (d *ParallelsDriver) Detect() *VMStatus {
	if runtime.GOOS != "darwin" {
		return &VMStatus{
			Type:       TypeParallels,
			Name:       d.Name(),
			Installed:  false,
			Path:       "",
			Version:    "Supported only on macOS",
			Priority:   d.Priority(),
			CanBootRaw: false,
		}
	}

	// 1. Check prlctl command in PATH
	if path, err := exec.LookPath("prlctl"); err == nil {
		return &VMStatus{
			Type:       TypeParallels,
			Name:       d.Name(),
			Installed:  true,
			Path:       path,
			Version:    "Parallels Desktop (prlctl CLI)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Check /Applications/Parallels Desktop.app bundle
	appPath := "/Applications/Parallels Desktop.app"
	if info, err := os.Stat(appPath); err == nil && info.IsDir() {
		return &VMStatus{
			Type:       TypeParallels,
			Name:       d.Name(),
			Installed:  true,
			Path:       appPath,
			Version:    "Parallels Desktop (/Applications/Parallels Desktop.app)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	return &VMStatus{
		Type:       TypeParallels,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *ParallelsDriver) Launch(ctx context.Context, diskPath string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	if os.Getenv("UNIBOOT_DRY_RUN") == "1" {
		logger.Info("UNIBOOT_DRY_RUN mode active, dry-run Parallels launch complete", "diskPath", diskPath)
		return nil
	}

	logger.Info("Executing Parallels Desktop preview simulation test", "disk", diskPath, "path", status.Path)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	ensureDiskPermissions(targetPath)

	cmd := exec.Command("open", "-a", "Parallels Desktop")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open Parallels Desktop application: %w", err)
	}

	return nil
}
