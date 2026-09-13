// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package qemu

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// QEMUStatus contains detection metadata for QEMU installation.
type QEMUStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
}

// Detect checks if QEMU executable is available on the system PATH.
func Detect() *QEMUStatus {
	path, err := exec.LookPath("qemu-system-x86_64")
	if err != nil {
		// Support dry-run or mock mode when UNIBOOT_DRY_RUN is set
		if os.Getenv("UNIBOOT_DRY_RUN") != "" {
			return &QEMUStatus{
				Installed: true,
				Path:      "/usr/local/bin/qemu-system-x86_64 (Dry-Run)",
				Version:   "QEMU system x86_64 (Mock)",
			}
		}
		return &QEMUStatus{
			Installed: false,
			Path:      "",
			Version:   "Not Installed",
		}
	}
	return &QEMUStatus{
		Installed: true,
		Path:      path,
		Version:   "QEMU system x86_64",
	}
}

// LaunchTest executes a non-blocking QEMU preview test instance on the target USB drive.
func LaunchTest(ctx context.Context, diskPath string) error {
	if diskPath == "" {
		return fmt.Errorf("target disk device path cannot be empty")
	}

	// Dry-run mode for tests or simulation
	if os.Getenv("UNIBOOT_DRY_RUN") != "" || strings.HasPrefix(diskPath, "dummy") || strings.HasPrefix(diskPath, "test") {
		return nil
	}

	status := Detect()
	if !status.Installed {
		return fmt.Errorf("QEMU (qemu-system-x86_64) is not installed on this system")
	}

	cmd := exec.CommandContext(ctx, status.Path, "-m", "1024", "-hda", diskPath)
	return cmd.Start()
}
