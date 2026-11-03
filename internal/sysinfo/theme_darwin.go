// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build darwin

package sysinfo

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// IsSystemDarkTheme returns true if macOS is currently configured in Dark Mode appearance.
func IsSystemDarkTheme() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "read", "-g", "AppleInterfaceStyle")
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

// SyncTitleBarTheme is a no-op on macOS as window appearance is handled natively by Cocoa / NSAppearance.
func SyncTitleBarTheme(theme string) {
}
