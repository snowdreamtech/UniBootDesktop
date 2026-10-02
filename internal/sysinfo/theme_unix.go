// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !darwin && !windows

package sysinfo

import (
	"os/exec"
	"strings"
)

// IsSystemDarkTheme returns true if Linux/Unix desktop is currently configured in Dark Mode.
func IsSystemDarkTheme() bool {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	out, err := cmd.Output()
	if err == nil && strings.Contains(strings.ToLower(string(out)), "dark") {
		return true
	}
	return false
}
