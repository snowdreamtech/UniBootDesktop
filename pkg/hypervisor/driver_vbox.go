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
	"strings"

	"github.com/snowdreamtech/unigodesktop/internal/logger"
)

type VirtualBoxDriver struct{}

func (d *VirtualBoxDriver) Type() HypervisorType {
	return TypeVirtualBox
}

func (d *VirtualBoxDriver) Name() string {
	return "Oracle VM VirtualBox"
}

func (d *VirtualBoxDriver) Priority() int {
	return 4
}

func (d *VirtualBoxDriver) Detect() *VMStatus {
	// 1. Check system PATH for VBoxManage
	if path, err := exec.LookPath("VBoxManage"); err == nil {
		return &VMStatus{
			Type:       TypeVirtualBox,
			Name:       d.Name(),
			Installed:  true,
			Path:       path,
			Version:    "VirtualBox (VBoxManage CLI)",
			Priority:   d.Priority(),
			CanBootRaw: true,
		}
	}

	// 2. Platform specific paths
	commonPaths := []string{}
	if runtime.GOOS == "darwin" {
		commonPaths = append(commonPaths,
			"/usr/local/bin/VBoxManage",
			"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
			"/Applications/VirtualBox.app",
		)
	} else if runtime.GOOS == "windows" {
		commonPaths = append(commonPaths,
			`C:\Program Files\Oracle\VirtualBox\VBoxManage.exe`,
			`C:\Program Files (x86)\Oracle\VirtualBox\VBoxManage.exe`,
			`C:\Program Files\Oracle\VirtualBox\VirtualBox.exe`,
		)
	} else if runtime.GOOS == "linux" {
		commonPaths = append(commonPaths,
			"/usr/bin/VBoxManage",
			"/usr/bin/virtualbox",
		)
	}

	for _, p := range commonPaths {
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() || strings.HasSuffix(p, ".app") {
				return &VMStatus{
					Type:       TypeVirtualBox,
					Name:       d.Name(),
					Installed:  true,
					Path:       p,
					Version:    fmt.Sprintf("VirtualBox (%s)", filepath.Base(p)),
					Priority:   d.Priority(),
					CanBootRaw: true,
				}
			}
		}
	}

	return &VMStatus{
		Type:       TypeVirtualBox,
		Name:       d.Name(),
		Installed:  false,
		Path:       "",
		Version:    "Not Installed",
		Priority:   d.Priority(),
		CanBootRaw: false,
	}
}

func (d *VirtualBoxDriver) Launch(ctx context.Context, diskPath string) error {
	status := d.Detect()
	if !status.Installed {
		return fmt.Errorf("%s is not installed on host system", d.Name())
	}

	logger.Info("Executing VirtualBox preview test instance", "disk", diskPath, "vboxPath", status.Path)

	targetPath := ResolveRawDiskDevice(diskPath)
	if targetPath == "" {
		targetPath = diskPath
	}
	ensureDiskPermissions(targetPath)

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "VirtualBox")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open VirtualBox application: %w", err)
		}
		return nil
	}

	if path, err := exec.LookPath("VBoxManage"); err == nil {
		cmd := exec.Command(path, "list", "vms")
		_ = cmd.Run()
		return nil
	}

	cmd := exec.Command(status.Path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start VirtualBox: %w", err)
	}

	return nil
}
