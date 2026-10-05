// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build darwin

package sysinfo

import (
	"os/exec"
	"strings"
)

// IsSystemDarkTheme returns true if macOS is currently configured in Dark Mode appearance.
func IsSystemDarkTheme() bool {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "Dark"
}

// SupportsMicaBackdrop always returns false on macOS.
func SupportsMicaBackdrop() bool {
	return false
}
