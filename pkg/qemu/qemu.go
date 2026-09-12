// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package qemu

import (
	"context"
	"fmt"
	"os/exec"
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
	status := Detect()
	if !status.Installed {
		return fmt.Errorf("QEMU is not installed on this system")
	}

	cmd := exec.CommandContext(ctx, status.Path, "-m", "1024", "-hda", diskPath)
	return cmd.Start()
}
