// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows && !darwin

package sysinfo

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// IsSystemDarkTheme returns true if Linux desktop environment is configured in Dark Mode.
func IsSystemDarkTheme() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "dark") {
		return true
	}
	return false
}

// SupportsMicaBackdrop always returns false on Linux.
func SupportsMicaBackdrop() bool {
	return false
}

// ApplyThemeToWindow is a no-op on Linux.
func ApplyThemeToWindow(hwnd uintptr, dark, mica bool) error {
	return nil
}

// SyncTitleBarTheme is a no-op on Linux as GTK window titlebars are handled by the window manager or Wails.
func SyncTitleBarTheme(theme string) {
}
